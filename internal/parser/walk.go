package parser

import (
	"fmt"
	"strings"

	ts "github.com/tree-sitter/go-tree-sitter"

	"github.com/infrashift/semscope/internal/assert"
)

// byteRange is a range of source bytes that the stripped text blanks out.
type byteRange struct {
	start uint
	end   uint
}

// walker holds the state of one tree walk. The Parser reuses it.
type walker struct {
	f       *File
	src     []byte
	lang    Language
	limits  Limits
	fnStack []int
	fnEnds  []uint
	blanks  []byteRange
	records int
}

// run visits each node of the tree under root in pre-order, without recursion.
func (w *walker) run(f *File, src []byte, root *ts.Node, limits Limits) error {
	assert.That(f != nil && root != nil, "f and root are not nil")
	w.f, w.src, w.lang, w.limits, w.records = f, src, f.Lang, limits, 0
	w.fnStack, w.fnEnds, w.blanks = w.fnStack[:0], w.fnEnds[:0], w.blanks[:0]
	c := root.Walk()
	defer c.Close()
	depth := 0
	for steps := 0; steps < limits.MaxNodes; steps++ {
		n := c.Node()
		w.popEnded(n.StartByte())
		descend, err := w.visit(n)
		if err != nil {
			return err
		}
		if descend && c.GotoFirstChild() {
			if depth++; depth > limits.MaxDepth {
				return fmt.Errorf("%s: the syntax tree is deeper than %d levels", f.Path, limits.MaxDepth)
			}
			continue
		}
		more, up := nextNode(c, depth)
		depth -= up
		if !more {
			w.f.Stripped = strip(w.src, w.blanks)
			assert.That(depth >= 0, "the walk ends at the root")
			return nil
		}
	}
	return fmt.Errorf("%s: more than %d syntax nodes", f.Path, limits.MaxNodes)
}

// nextNode moves c to the next node in pre-order that is not a descendant of
// the current node. It returns false at the end of the tree, and the number
// of levels that it went up.
func nextNode(c *ts.TreeCursor, depth int) (bool, int) {
	assert.That(depth >= 0, "the depth is not negative")
	up := 0
	for ; up <= depth; up++ {
		if c.GotoNextSibling() {
			return true, up
		}
		if !c.GotoParent() {
			return false, up
		}
	}
	assert.That(up > depth, "the loop ends above the root")
	return false, up
}

// popEnded removes the functions that end before the byte offset start.
func (w *walker) popEnded(start uint) {
	assert.That(len(w.fnStack) == len(w.fnEnds), "the function stacks have the same size")
	for i := len(w.fnStack); i > 0 && start >= w.fnEnds[i-1]; i-- {
		w.fnStack = w.fnStack[:i-1]
		w.fnEnds = w.fnEnds[:i-1]
	}
	assert.That(len(w.fnStack) == len(w.fnEnds), "the stacks shrink together")
}

// visit records node n. It returns false when the walk must skip the children.
func (w *walker) visit(n *ts.Node) (bool, error) {
	assert.That(n != nil, "n is not nil")
	assert.That(w.f != nil, "run sets the file")
	kind := n.Kind()
	switch {
	case kind == "comment" || kind == "line_comment" || kind == "block_comment":
		w.blank(n)
		if w.lang == Go {
			return false, w.goComment(n)
		}
		return false, nil
	case isStringKind(kind):
		w.blank(n)
		return false, nil
	}
	switch w.lang {
	case Go:
		return w.visitGo(n, kind)
	case C, CPP:
		return w.visitC(n, kind)
	case Zig:
		return w.visitZig(n, kind)
	case Python:
		return w.visitPython(n, kind)
	case Rust:
		return w.visitRust(n, kind)
	}
	return false, fmt.Errorf("%s: unknown language %q", w.f.Path, w.lang)
}

func isStringKind(kind string) bool {
	switch kind {
	case "interpreted_string_literal", "raw_string_literal", "rune_literal", "string_literal",
		"char_literal", "system_lib_string", "string", "multiline_string", "character", "concatenated_string":
		return true
	}
	return false
}

// strip returns the lines of src with each blank range replaced by spaces.
// It keeps the newlines, so the line numbers stay the same.
func strip(src []byte, blanks []byteRange) []string {
	out := []byte(string(src))
	for _, r := range blanks {
		assert.That(r.start <= r.end, "a blank range is ordered")
		for i := r.start; i < r.end && i < uint(len(out)); i++ {
			if out[i] != '\n' {
				out[i] = ' '
			}
		}
	}
	lines := splitLines(out)
	assert.That(len(out) == len(src), "stripping keeps the size")
	return lines
}

func (w *walker) blank(n *ts.Node) {
	assert.That(n.StartByte() <= n.EndByte(), "a node range is ordered")
	assert.That(n.EndByte() <= uint(len(w.src)), "the node lies inside the source")
	w.blanks = append(w.blanks, byteRange{start: n.StartByte(), end: n.EndByte()})
}

// text returns the source text of n.
func (w *walker) text(n *ts.Node) string {
	assert.That(n != nil, "n is not nil")
	assert.That(n.EndByte() <= uint(len(w.src)), "the node lies inside the source")
	return n.Utf8Text(w.src)
}

// enclosing returns the index of the innermost function, or -1.
func (w *walker) enclosing() int {
	assert.That(len(w.fnStack) == len(w.fnEnds), "the function stacks have the same size")
	assert.That(len(w.fnStack) <= len(w.f.Functions), "each stacked function exists")
	if len(w.fnStack) == 0 {
		return -1
	}
	return w.fnStack[len(w.fnStack)-1]
}

// NodePos returns the 1-based range of node n.
func NodePos(n *ts.Node) Pos {
	s, e := n.StartPosition(), n.EndPosition()
	p := Pos{Line: int(s.Row) + 1, Col: int(s.Column) + 1, EndLine: int(e.Row) + 1, EndCol: int(e.Column) + 1}
	assert.That(p.Line >= 1 && p.Col >= 1, "positions are 1-based")
	assert.That(p.EndLine >= p.Line, "a range ends after it starts")
	return p
}

func pos(n *ts.Node) Pos {
	return NodePos(n)
}

// count adds one record and checks the record limit.
func (w *walker) count() error {
	assert.That(w.limits.MaxRecords > 0, "New checks the limits")
	if w.records++; w.records > w.limits.MaxRecords {
		return fmt.Errorf("%s: more than %d records", w.f.Path, w.limits.MaxRecords)
	}
	assert.That(w.records >= 1, "the count includes this record")
	return nil
}

// addFunction records the function fn of node n and makes it the innermost function.
func (w *walker) addFunction(n *ts.Node, fn Function) error {
	assert.That(fn.Name != "", "a function has a name")
	assert.That(n != nil, "n is not nil")
	if err := w.count(); err != nil {
		return err
	}
	if len(w.fnStack) >= w.limits.MaxNesting {
		return fmt.Errorf("%s:%d: functions nest deeper than %d levels", w.f.Path, pos(n).Line, w.limits.MaxNesting)
	}
	if fn.Key == "" {
		fn.Key = fn.Name
	}
	fn.Pos = pos(n)
	w.fnStack = append(w.fnStack, len(w.f.Functions))
	w.fnEnds = append(w.fnEnds, n.EndByte())
	w.f.Functions = append(w.f.Functions, fn)
	assert.That(w.enclosing() == len(w.f.Functions)-1, "the new function is the innermost function")
	return nil
}

// addCall records a call of the callee text at node n.
func (w *walker) addCall(n *ts.Node, callee string) error {
	assert.That(n != nil, "n is not nil")
	assert.That(callee != "", "a call has a callee")
	if err := w.count(); err != nil {
		return err
	}
	callee = normalizeCallee(callee)
	qual, name := splitCallee(callee)
	assert.That(strings.HasSuffix(callee, name), "the name ends the callee")
	parent := n.Parent()
	unused := parent != nil && parent.Kind() == "expression_statement"
	w.f.Calls = append(w.f.Calls, Call{
		Fn: w.enclosing(), Callee: callee, Name: name, Qual: qual, Unused: unused, Pos: pos(n),
	})
	return nil
}

// addSite records a site of kind k at node n.
func (w *walker) addSite(n *ts.Node, k SiteKind, detail string, depth int) error {
	assert.That(k >= SiteGoto && k <= SiteFuncPointer, "the site kind is known")
	assert.That(depth >= 0, "the depth is not negative")
	if err := w.count(); err != nil {
		return err
	}
	w.f.Sites = append(w.f.Sites, Site{Kind: k, Detail: detail, Depth: depth, Fn: w.enclosing(), Pos: pos(n)})
	return nil
}

// normalizeCallee removes white space and type arguments in [] or <> from a callee text.
func normalizeCallee(s string) string {
	var b strings.Builder
	depth := 0
	for _, r := range strings.ReplaceAll(s, "->", ".") {
		switch {
		case r == '[' || r == '<':
			depth++
		case r == ']' || r == '>':
			depth = max(depth-1, 0)
		case depth == 0 && r != ' ' && r != '\t' && r != '\n' && r != '\r':
			b.WriteRune(r)
		}
	}
	out := strings.TrimSuffix(strings.ReplaceAll(b.String(), "::::", "::"), "::")
	assert.That(len(out) <= len(s), "normalizing does not grow the text")
	assert.That(!strings.ContainsAny(out, " \t\n"), "no white space remains")
	return out
}

// splitCallee splits a callee text at the last "." or "::" into a qualifier and a name.
func splitCallee(callee string) (string, string) {
	cut := strings.LastIndexAny(callee, ".:")
	if cut < 0 {
		return "", callee
	}
	name := callee[cut+1:]
	head := strings.TrimRight(callee[:cut], ":.")
	qcut := strings.LastIndexAny(head, ".:")
	assert.That(qcut < len(head), "the qualifier cut is inside the head")
	assert.That(strings.HasSuffix(callee, name), "the name ends the callee")
	return head[qcut+1:], name
}
