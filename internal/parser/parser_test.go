package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// summary prints the record of f in a stable text form for golden files.
func summary(f *File) string {
	var b strings.Builder
	fmt.Fprintf(&b, "package=%q imports=%v exempt=%v lines=%d\n", f.Package, f.Imports, f.Exempt, len(f.Lines))
	for _, fn := range f.Functions {
		fmt.Fprintf(&b, "fn %s key=%s recv=%s lines=%d-%d exempt=%v\n", fn.Name, fn.Key, fn.Recv, fn.Pos.Line, fn.Pos.EndLine, fn.Exempt)
	}
	for _, c := range f.Calls {
		fmt.Fprintf(&b, "call fn=%d %q name=%q qual=%q unused=%v line=%d\n", c.Fn, c.Callee, c.Name, c.Qual, c.Unused, c.Pos.Line)
	}
	for _, s := range f.Sites {
		fmt.Fprintf(&b, "site kind=%d %q depth=%d fn=%d line=%d\n", s.Kind, s.Detail, s.Depth, s.Fn, s.Pos.Line)
	}
	return b.String()
}

func newTestParser(t *testing.T) *Parser {
	t.Helper()
	p, err := New(DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(p.Close)
	return p
}

func TestSamplesGolden(t *testing.T) {
	p := newTestParser(t)
	all, err := filepath.Glob("testdata/sample_*")
	paths := []string{}
	for _, path := range all {
		if !strings.HasSuffix(path, ".golden") {
			paths = append(paths, path)
		}
	}
	if err != nil || len(paths) != 7 {
		t.Fatalf("samples = %v, %v", paths, err)
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) { checkSampleGolden(t, p, path) })
	}
}

// checkSampleGolden parses one sample and compares its summary with the golden file.
func checkSampleGolden(t *testing.T, p *Parser, path string) {
	t.Helper()
	src, err := os.ReadFile(path) //nolint:gosec // fixed testdata path
	if err != nil {
		t.Fatal(err)
	}
	lang, ok := Detect(path)
	if !ok {
		t.Fatalf("Detect(%s) failed", path)
	}
	f, _, err := p.Parse(path, lang, src)
	if err != nil {
		t.Fatal(err)
	}
	got := summary(f)
	golden := path + ".golden"
	if os.Getenv("SEMSCOPE_UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(golden, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden) //nolint:gosec // fixed testdata path
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("record differs from %s:\n--- got\n%s--- want\n%s", golden, got, want)
	}
}

func TestStrippedBlanksCommentsAndStrings(t *testing.T) {
	p := newTestParser(t)
	src := "package a\n\n// goto x\nvar s = \"for {\" // tail\nvar r = `\nfor {\n`\n"
	f, _, err := p.Parse("a.go", Go, []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(f.Stripped, "\n")
	if strings.Contains(joined, "goto") || strings.Contains(joined, "for") || len(f.Stripped) != len(f.Lines) {
		t.Fatalf("stripped = %q", f.Stripped)
	}
	if !strings.Contains(f.Stripped[3], "var s =") {
		t.Fatalf("stripped line 4 = %q", f.Stripped[3])
	}
}

func TestTestFilesAreExempt(t *testing.T) {
	p := newTestParser(t)
	for _, tc := range []struct {
		path string
		lang Language
	}{{"x_test.go", Go}, {"test_x.py", Python}, {"x_test.py", Python}} {
		f, tree, err := p.Parse(tc.path, tc.lang, []byte("x"))
		if err != nil || tree != nil || !f.Exempt {
			t.Errorf("Parse(%s) = exempt %v, tree %v, err %v", tc.path, f != nil && f.Exempt, tree != nil, err)
		}
	}
}

func TestCallsHaveNames(t *testing.T) {
	p := newTestParser(t)
	src := "fn f(n: usize) void {\n    for ((n - 1)..(2 * n)) |_| {}\n    for (0..(n - 1) / 2) |_| {}\n    g();\n}\n"
	f, _, err := p.Parse("range.zig", Zig, []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range f.Calls {
		if c.Name == "" {
			t.Errorf("call without a name: callee %q at line %d", c.Callee, c.Pos.Line)
		}
	}
}

func TestZig017Identifiers(t *testing.T) {
	p := newTestParser(t)
	src := "fn cancel(ev: *Ev) void {\n    const usingnamespace_x = 1;\n    await(ev, usingnamespace_x);\n    async(ev);\n}\n\nfn next() void {}\n"
	f, tree, err := p.Parse("io.zig", Zig, []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if tree.RootNode().HasError() {
		t.Fatalf("Zig 0.17 identifiers give a parse error: %s", tree.RootNode().ToSexp())
	}
	if len(f.Functions) != 2 || f.Functions[0].Pos.EndLine != 5 {
		t.Fatalf("functions = %+v, want cancel on lines 1-5 and next", f.Functions)
	}
	var names []string
	for _, c := range f.Calls {
		names = append(names, c.Name)
	}
	if strings.Join(names, ",") != "await,async" {
		t.Errorf("calls = %v, want await and async", names)
	}
}

func TestZigPrefixCalleesAndCommentBlocks(t *testing.T) {
	p := newTestParser(t)
	src := "fn f(T: type) void {\n    _ = ?Span(T);\n    if (!valid(T)) {}\n    const q: *Q(T) = undefined;\n    g() catch {\n        // ignored\n    };\n}\n"
	f, _, err := p.Parse("p.zig", Zig, []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, c := range f.Calls {
		names = append(names, c.Name)
	}
	if strings.Join(names, ",") != "Span,valid,Q,g" {
		t.Errorf("calls = %v, want Span, valid, Q and g", names)
	}
	swallowed := 0
	for _, s := range f.Sites {
		if s.Kind == SiteDiscard && strings.Contains(s.Detail, "catch {}") {
			swallowed++
		}
	}
	if swallowed != 1 {
		t.Errorf("a catch block with a comment only gave %d findings, want 1", swallowed)
	}
}

func TestZigAsmStructClobbers(t *testing.T) {
	p := newTestParser(t)
	src := "pub fn a() u64 {\n    return asm volatile (\"ecall\"\n        : [ret] \"={x10}\" (-> u64),\n        : [number] \"{x17}\" (1),\n        : .{\n          .memory = true,\n        });\n}\n\npub fn b() void {}\n"
	f, tree, err := p.Parse("asm.zig", Zig, []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if tree.RootNode().HasError() || len(f.Functions) != 2 {
		t.Fatalf("struct clobbers: error %v, %d functions, want no error and 2: %s",
			tree.RootNode().HasError(), len(f.Functions), tree.RootNode().ToSexp())
	}
}

// TestZigEmptyContainersAndRareForms checks the four changes that the grammar
// takes from ztree (ADR-0012): empty containers, an error set as a container
// field, a slice of an if type and an identifier as the asm clobbers.
func TestZigEmptyContainersAndRareForms(t *testing.T) {
	p := newTestParser(t)
	src := "const A = struct {};\nconst B = enum {};\nconst C = union {};\nconst D = *opaque {};\n" +
		"const P = struct { w: error{A}, v: error{ B, C } = undefined };\n" +
		"const T = struct { error{foo}, u8 };\n" +
		"const S = struct { vec: []const if (c) A else B };\n" +
		"fn g() void {\n    asm volatile (\"svc #0\"\n        : [r] \"={x0}\" (-> usize),\n        : [n] \"{x8}\" (n),\n        : clobbers);\n}\n\n" +
		"fn h() void {}\n"
	f, tree, err := p.Parse("forms.zig", Zig, []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if tree.RootNode().HasError() || len(f.Functions) != 2 {
		t.Fatalf("rare forms: error %v, %d functions, want no error and 2: %s",
			tree.RootNode().HasError(), len(f.Functions), tree.RootNode().ToSexp())
	}
}

func TestLimits(t *testing.T) {
	small := DefaultLimits()
	small.MaxFileBytes = 4
	p, err := New(small)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	if _, _, err := p.Parse("a.go", Go, []byte("package a")); err == nil || !strings.Contains(err.Error(), "max 4") {
		t.Fatalf("file size limit: err = %v", err)
	}
	deep := DefaultLimits()
	deep.MaxDepth = 3
	q, err := New(deep)
	if err != nil {
		t.Fatal(err)
	}
	defer q.Close()
	if _, _, err := q.Parse("a.go", Go, []byte("package a\nfunc f() { g(h(i(j(1)))) }\n")); err == nil || !strings.Contains(err.Error(), "deeper than 3") {
		t.Fatalf("depth limit: err = %v", err)
	}
	if _, err := New(Limits{}); err == nil {
		t.Fatalf("New(Limits{}) gave no error")
	}
}

func TestDetect(t *testing.T) {
	for path, want := range map[string]Language{"a.go": Go, "a.h": C, "a.HPP": CPP, "a.zig": Zig, "a.py": Python, "a.rs": Rust} {
		if got, ok := Detect(path); !ok || got != want {
			t.Errorf("Detect(%s) = %s, %v", path, got, ok)
		}
	}
	if _, ok := Detect("a.txt"); ok {
		t.Errorf("Detect(a.txt) = ok")
	}
}

func TestNormalizeAndSplitCallee(t *testing.T) {
	for in, want := range map[string][3]string{
		"assert.That":      {"assert.That", "assert", "That"},
		"a -> b":           {"a.b", "a", "b"},
		"Vec::<u8>::new":   {"Vec::new", "Vec", "new"},
		"f[int]":           {"f", "", "f"},
		"std.debug.assert": {"std.debug.assert", "debug", "assert"},
		"p[0].Msg":         {"p[].Msg", "p[]", "Msg"},
		".init":            {".init", ".", "init"},
		"::f":              {"::f", "", "f"},
		"x.items[i][j].f":  {"x.items[][].f", "items[][]", "f"},
	} {
		got := normalizeCallee(in)
		qual, name := splitCallee(got)
		if got != want[0] || qual != want[1] || name != want[2] {
			t.Errorf("%q -> %q (%q, %q), want %v", in, got, qual, name, want)
		}
	}
}
