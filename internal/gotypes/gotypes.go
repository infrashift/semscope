// Package gotypes loads Go packages with type information and finds each
// call that drops an error result.
//
// The Tree-sitter adapter of Go has no type information, so it finds only
// the explicit discards `_ = f()` and `x, _ := f()`. This package finds the
// call statements, the defer statements and the go statements whose call
// returns an error. The rule NASA-P10-07 uses it when its parameter
// go_types is true. ADR-0009 records the decision.
package gotypes

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/infrashift/semscope/internal/assert"
)

// MaxModuleDepth bounds the search for go.mod above a directory.
const MaxModuleDepth = 256

// Statement kinds of a Drop.
const (
	StmtCall  = "call"
	StmtDefer = "defer"
	StmtGo    = "go"
)

// Drop is one call that drops an error result.
type Drop struct {
	// File is the absolute path of the source file, as Find received it.
	File    string
	Line    int
	Col     int
	EndLine int
	EndCol  int
	// Callee is the full name of the called function, for example
	// "os.Remove" or "(*os.File).Close". A call of a function value gives
	// the text of the expression.
	Callee string
	// Arg is the type of the first argument, else "". The ignore entry
	// "fmt.Fprintf(*strings.Builder)" uses it.
	Arg string
	// Stmt is StmtCall, StmtDefer or StmtGo.
	Stmt string
}

// Find loads the packages of the Go files in files and returns each drop
// in these files, sorted by file and position. Each file path is absolute.
// Find groups the files by the nearest go.mod and runs one load for each
// module. More than maxPackages packages give an error.
func Find(files []string, maxPackages int) ([]Drop, error) {
	assert.That(maxPackages > 0, "the package limit is positive")
	byModule, wanted, err := groupByModule(files)
	if err != nil {
		return nil, err
	}
	modules := make([]string, 0, len(byModule))
	for m := range byModule {
		modules = append(modules, m)
	}
	sort.Strings(modules)
	var out []Drop
	loaded := 0
	for _, m := range modules {
		drops, n, err := loadModule(m, byModule[m], wanted, maxPackages-loaded)
		if err != nil {
			return nil, err
		}
		loaded += n
		out = append(out, drops...)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		switch {
		case a.File != b.File:
			return a.File < b.File
		case a.Line != b.Line:
			return a.Line < b.Line
		}
		return a.Col < b.Col
	})
	assert.That(loaded <= maxPackages, "loadModule enforces the package limit")
	return out, nil
}

// groupByModule maps each module root to the package patterns of files. It
// also maps the resolved path of each file to the path that the caller gave.
func groupByModule(files []string) (map[string][]string, map[string]string, error) {
	byModule := map[string][]string{}
	wanted := make(map[string]string, len(files))
	seen := map[string]bool{}
	for _, f := range files {
		assert.That(filepath.IsAbs(f), "the engine gives absolute paths")
		resolved, err := filepath.EvalSymlinks(f)
		if err != nil {
			return nil, nil, fmt.Errorf("gotypes: %w", err)
		}
		wanted[resolved] = f
		dir := filepath.Dir(resolved)
		if seen[dir] {
			continue
		}
		seen[dir] = true
		root, err := moduleRoot(dir)
		if err != nil {
			return nil, nil, err
		}
		rel, err := filepath.Rel(root, dir)
		if err != nil {
			return nil, nil, fmt.Errorf("gotypes: %w", err)
		}
		byModule[root] = append(byModule[root], "./"+filepath.ToSlash(rel))
	}
	assert.That(len(wanted) <= len(files), "one entry for each file")
	return byModule, wanted, nil
}

// moduleRoot returns the nearest directory at or above dir that holds go.mod.
func moduleRoot(dir string) (string, error) {
	assert.That(filepath.IsAbs(dir), "dir is absolute")
	d := dir
	for i := 0; i < MaxModuleDepth; i++ {
		if st, err := os.Stat(filepath.Join(d, "go.mod")); err == nil && !st.IsDir() {
			return d, nil
		}
		parent := filepath.Dir(d)
		if parent == d {
			break
		}
		d = parent
	}
	assert.That(d != "", "the search ends at a directory")
	return "", fmt.Errorf("gotypes: %s: no go.mod in the directory or a parent directory", dir)
}

// loadModule loads the packages of patterns in the module at root and
// returns the drops in the wanted files and the number of packages.
func loadModule(root string, patterns []string, wanted map[string]string, maxPackages int) ([]Drop, int, error) {
	assert.That(len(patterns) > 0, "a module has one or more packages")
	sort.Strings(patterns)
	fset := token.NewFileSet()
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:  root,
		Fset: fset,
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, 0, fmt.Errorf("gotypes: load %s: %w", root, err)
	}
	if len(pkgs) > maxPackages {
		return nil, 0, fmt.Errorf("gotypes: more than %d packages", maxPackages)
	}
	w := &walker{fset: fset, wanted: wanted, errType: types.Universe.Lookup("error").Type()}
	for _, p := range pkgs {
		if err := loadError(p); err != nil {
			return nil, 0, err
		}
		w.info = p.TypesInfo
		for _, f := range p.Syntax {
			w.file(f)
		}
	}
	assert.That(w.errType != nil, "the universe has the error type")
	return w.out, len(pkgs), nil
}

// loadError gives the first error of p. A package whose build constraints
// exclude every file has no code to check, so it is not an error.
func loadError(p *packages.Package) error {
	assert.That(p != nil, "p is not nil")
	for _, e := range p.Errors {
		if strings.Contains(e.Msg, "build constraints exclude all Go files") {
			continue
		}
		return fmt.Errorf("gotypes: package %s: %s", p.PkgPath, e.Error())
	}
	assert.That(p.TypesInfo != nil || len(p.Syntax) == 0, "a package with syntax has type information")
	return nil
}

// walker finds the drops in the syntax trees of one module.
type walker struct {
	fset    *token.FileSet
	info    *types.Info
	wanted  map[string]string
	errType types.Type
	out     []Drop
}

// file records the drops in one syntax tree. ast.Inspect visits each node.
func (w *walker) file(f *ast.File) {
	assert.That(f != nil && w.info != nil, "the package has syntax and type information")
	if _, ok := w.wanted[w.fset.Position(f.Pos()).Filename]; !ok {
		return
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.ExprStmt:
			if call, ok := ast.Unparen(s.X).(*ast.CallExpr); ok {
				w.check(call, StmtCall)
			}
		case *ast.DeferStmt:
			w.check(s.Call, StmtDefer)
		case *ast.GoStmt:
			w.check(s.Call, StmtGo)
		}
		return true
	})
	assert.That(w.fset != nil, "the walker has a file set")
}

// check records call when it returns an error.
func (w *walker) check(call *ast.CallExpr, stmt string) {
	assert.That(call != nil, "a statement holds a call")
	tv, ok := w.info.Types[call]
	if !ok || tv.IsType() || !w.returnsError(tv.Type) {
		return
	}
	pos, end := w.fset.Position(call.Pos()), w.fset.Position(call.End())
	d := Drop{
		File: w.wanted[pos.Filename], Line: pos.Line, Col: pos.Column, EndLine: end.Line, EndCol: end.Column,
		Callee: types.ExprString(call.Fun), Stmt: stmt,
	}
	if fn, ok := typeutil.Callee(w.info, call).(*types.Func); ok {
		d.Callee = fn.Origin().FullName()
	}
	if len(call.Args) > 0 {
		if t := w.info.TypeOf(call.Args[0]); t != nil {
			d.Arg = types.TypeString(t, nil)
		}
	}
	assert.That(d.File != "", "the walker visits wanted files only")
	w.out = append(w.out, d)
}

// returnsError reports whether a result of type t is an error. A tuple
// holds the results of a call with more than one result.
func (w *walker) returnsError(t types.Type) bool {
	assert.That(t != nil, "a call has a type")
	assert.That(w.errType != nil, "loadModule sets the error type")
	tuple, ok := t.(*types.Tuple)
	if !ok {
		return w.isError(t)
	}
	for i := range tuple.Len() {
		if w.isError(tuple.At(i).Type()) {
			return true
		}
	}
	return false
}

// isError reports whether t is the error interface or an interface that
// includes it.
func (w *walker) isError(t types.Type) bool {
	assert.That(t != nil, "a result has a type")
	if types.Identical(t, w.errType) {
		return true
	}
	iface, ok := w.errType.Underlying().(*types.Interface)
	assert.That(ok, "error is an interface")
	return types.IsInterface(t) && types.Implements(t, iface)
}

// Ignored reports whether an entry of ignore names the callee of d. An
// entry is a full function name, for example "(*strings.Builder).WriteString",
// or a full name and the type of the first argument, for example
// "fmt.Fprintf(*strings.Builder)".
func (d Drop) Ignored(ignore []string) bool {
	assert.That(d.Callee != "", "a drop names its callee")
	assert.That(d.Stmt != "", "a drop has a statement kind")
	withArg := d.Callee + "(" + d.Arg + ")"
	for _, entry := range ignore {
		if entry == d.Callee || (d.Arg != "" && entry == withArg) {
			return true
		}
	}
	return false
}
