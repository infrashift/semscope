package engine

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/infrashift/semscope/internal/assert"
	"github.com/infrashift/semscope/internal/parser"
)

// graphNode is one function in the call graph. Functions with the same key
// share one node, the same as in the oracles.
type graphNode struct {
	key   string
	file  *parser.File
	fn    int
	edges []int
}

// callGraph is a call graph with a fixed upper bound on its node count.
type callGraph struct {
	nodes []graphNode
	index map[string]int
	edges int
}

// globalChecks runs the checks over all files: recursion and the assertion average.
func (e *Engine) globalChecks(files []*parser.File) ([]Finding, map[parser.Language]LangStats, error) {
	assert.That(len(files) <= e.limits.MaxFiles, "collect enforces MaxFiles")
	var out []Finding
	stats := map[parser.Language]LangStats{}
	for i := range e.rules {
		cr := &e.rules[i]
		switch cr.rule.Check.Kind {
		case "no_goto_or_recursion":
			g, err := buildGraph(cr, files, e.limits.MaxGraphNodes)
			if err != nil {
				return nil, nil, err
			}
			out = append(out, cycleFindings(cr, g)...)
		case "min_assertions":
			found, s := assertionAverages(cr, files)
			assert.That(len(found) <= len(parser.Languages()), "one average finding for each language at most")
			out, stats = append(out, found...), s
		}
	}
	assert.That(stats != nil, "the stats map exists")
	return out, stats, nil
}

// scope returns the name space of the functions of f: the package directory
// for Go, one shared space for C and C++, else the file.
func scope(f *parser.File) string {
	assert.That(f.Path != "", "the file has a path")
	assert.That(f.Lang != "", "the file has a language")
	switch f.Lang {
	case parser.Go:
		return path.Dir(f.Path)
	case parser.C, parser.CPP:
		return "<c>"
	}
	return f.Path
}

// resolve returns the node key of the function that call names, or "".
// It follows the edge rules of the oracles. An edge comes from a plain call,
// a call through the receiver or self, or a call through a Zig import name.
func resolve(f *parser.File, caller parser.Function, call parser.Call) string {
	assert.That(call.Name != "", "a call has a name")
	assert.That(caller.Key != "", "addFunction sets the key")
	sc := scope(f)
	switch {
	case f.Lang == parser.CSharp:
		return sc + "\x00" + csharpTarget(caller.Key, call)
	case call.Qual == "":
		return sc + "\x00" + call.Name
	case f.Lang == parser.Go && caller.Recv != "" && call.Qual == caller.Recv:
		recvType, _, found := strings.Cut(caller.Key, ".")
		if !found {
			return ""
		}
		return sc + "\x00" + recvType + "." + call.Name
	case f.Lang == parser.Python && (call.Qual == "self" || call.Qual == "cls"):
		return sc + "\x00" + call.Name
	case f.Lang == parser.Rust && (call.Qual == "self" || call.Qual == "Self"):
		return sc + "\x00" + call.Name
	}
	if imp, ok := f.Imports[call.Qual]; ok && f.Lang == parser.Zig && strings.HasSuffix(imp, ".zig") {
		return path.Join(path.Dir(f.Path), imp) + "\x00" + call.Name
	}
	return ""
}

// csharpTarget returns the key of the C# function that call names. A plain
// call and a call through `this` name a function of the type of the caller.
// A call through a type name, such as `Parser.Parse`, names a function of
// that type. A call through `base` or a variable gives "".
func csharpTarget(callerKey string, call parser.Call) string {
	assert.That(callerKey != "", "addFunction sets the key")
	assert.That(call.Name != "", "a call has a name")
	if call.Qual != "" && call.Qual != "this" {
		if call.Qual == "base" || strings.HasSuffix(call.Qual, "?") {
			return ""
		}
		return call.Qual + "." + call.Name
	}
	cut := strings.Index(callerKey, ".")
	if cut < 0 {
		return call.Name
	}
	return callerKey[:cut+1] + call.Name
}

// buildGraph builds the call graph of the files of rule cr. A test function has no node.
func buildGraph(cr *compiledRule, files []*parser.File, maxNodes int) (*callGraph, error) {
	assert.That(maxNodes > 0, "New checks the limits")
	g := &callGraph{index: map[string]int{}}
	for _, f := range files {
		if f.Exempt || !cr.langs[f.Lang] {
			continue
		}
		for i, fn := range f.Functions {
			key := scope(f) + "\x00" + fn.Key
			if _, dup := g.index[key]; fn.Exempt || dup {
				continue
			}
			if len(g.nodes) >= maxNodes {
				return nil, fmt.Errorf("engine: more than %d call graph nodes", maxNodes)
			}
			g.index[key] = len(g.nodes)
			g.nodes = append(g.nodes, graphNode{key: key, file: f, fn: i})
		}
	}
	for _, f := range files {
		if !f.Exempt && cr.langs[f.Lang] {
			g.addEdges(f)
		}
	}
	g.sortEdges()
	assert.That(len(g.nodes) == len(g.index), "one node for each key")
	return g, nil
}

// addEdges adds the resolved calls of f to the graph.
func (g *callGraph) addEdges(f *parser.File) {
	assert.That(!f.Exempt, "an exempt file has no edges")
	for _, call := range f.Calls {
		if call.Fn < 0 || f.Functions[call.Fn].Exempt {
			continue
		}
		caller := f.Functions[call.Fn]
		from, ok := g.index[scope(f)+"\x00"+caller.Key]
		to, found := g.target(f, caller, call)
		if ok && found {
			g.nodes[from].edges = append(g.nodes[from].edges, to)
		}
	}
	assert.That(g.index != nil, "the graph has an index")
}

// target returns the node of the function that call names. A plain Zig
// call and a Zig call through self, Self or @This() go through zigTarget. Other calls go
// through resolve.
func (g *callGraph) target(f *parser.File, caller parser.Function, call parser.Call) (int, bool) {
	assert.That(call.Name != "", "a call has a name")
	self := call.Qual == "self" || call.Qual == "Self" || call.Qual == "@This()"
	if f.Lang == parser.Zig && (call.Qual == "" || self) {
		return g.zigTarget(f, caller, call)
	}
	to, found := g.index[resolve(f, caller, call)]
	assert.That(!found || to < len(g.nodes), "the index holds valid nodes")
	return to, found
}

// zigTarget resolves a Zig call through the scopes of the caller, the
// lexical scope of Zig. A plain call `f()` looks in the container of the
// caller, then in each outer container, then in the file. A call through
// self looks in the container of the caller only.
func (g *callGraph) zigTarget(f *parser.File, caller parser.Function, call parser.Call) (int, bool) {
	assert.That(f.Lang == parser.Zig, "the file is Zig")
	var path []string
	if caller.Scope != "" {
		path = strings.Split(caller.Scope, ".")
	}
	lowest := 0
	if call.Qual != "" {
		lowest = len(path)
	}
	sc := scope(f) + "\x00"
	for i := len(path); i >= lowest; i-- {
		key := strings.Join(append(append([]string{}, path[:i]...), call.Name), ".")
		if to, found := g.index[sc+key]; found {
			return to, true
		}
	}
	assert.That(lowest <= len(path), "the search starts inside the path")
	return 0, false
}

// sortEdges sorts the edges of each node by target key and removes duplicates.
func (g *callGraph) sortEdges() {
	assert.That(len(g.nodes) == len(g.index), "one node for each key")
	before := 0
	for i := range g.nodes {
		before += len(g.nodes[i].edges)
	}
	g.edges = 0
	for i := range g.nodes {
		edges := g.nodes[i].edges
		sort.Slice(edges, func(a, b int) bool { return g.nodes[edges[a]].key < g.nodes[edges[b]].key })
		out := edges[:0]
		for j, to := range edges {
			if j == 0 || to != edges[j-1] {
				out = append(out, to)
			}
		}
		g.nodes[i].edges = out
		g.edges += len(out)
	}
	assert.That(g.edges <= before, "removing duplicates adds no edges")
}
