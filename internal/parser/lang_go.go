package parser

import (
	"strings"

	ts "github.com/infrashift/ztree/bindings/go"

	"github.com/infrashift/semscope/internal/assert"
)

// visitGo records the Go constructs of node n.
func (w *walker) visitGo(n *ts.Node, kind string) (bool, error) {
	assert.That(kind == n.Kind(), "visit passes the kind of n")
	assert.That(w.lang == Go, "the file is Go")
	switch kind {
	case "package_clause":
		if id := n.NamedChild(0); id != nil {
			w.f.Package = w.text(id)
		}
	case "function_declaration", "method_declaration":
		return true, w.goFunction(n)
	case "call_expression":
		if fn := n.ChildByFieldName("function"); fn != nil {
			return true, w.addCall(n, w.text(fn))
		}
	case "goto_statement":
		return true, w.addSite(n, SiteGoto, "`goto` statement", 0)
	case "for_statement":
		if goLoopUnbounded(n, w.src) {
			return true, w.addSite(n, SiteUnboundedLoop, "`for` loop without a condition", 0)
		}
	case "assignment_statement", "short_var_declaration", "var_spec":
		return true, w.goDiscard(n)
	case "var_declaration":
		return true, w.goGlobal(n)
	case "pointer_type":
		return w.pointerChain(n, "pointer_type", "pointer_type")
	case "function_type":
		return true, w.addSite(n, SiteFuncPointer, "function type", 0)
	}
	return true, nil
}

// goFunction records a Go function or method declaration.
func (w *walker) goFunction(n *ts.Node) error {
	name := n.ChildByFieldName("name")
	if name == nil || n.ChildByFieldName("body") == nil {
		return nil
	}
	fn := Function{Name: w.text(name)}
	fn.Exempt = isTestHelper(fn.Name)
	if recv := n.ChildByFieldName("receiver"); recv != nil {
		fn.Recv, fn.Key = w.goReceiver(recv, fn.Name)
	}
	assert.That(fn.Name != "", "a function has a name")
	assert.That(fn.Recv == "" || fn.Key != "", "a method has a key")
	return w.addFunction(n, fn)
}

// goReceiver returns the receiver variable and the call graph key
// "Type.name" of a method. The variable is "" when the receiver has no name.
func (w *walker) goReceiver(recv *ts.Node, name string) (string, string) {
	assert.That(recv.Kind() == "parameter_list", "a receiver is a parameter list")
	param := recv.NamedChild(0)
	if param == nil {
		return "", ""
	}
	variable, key := "", ""
	if v := param.ChildByFieldName("name"); v != nil {
		variable = w.text(v)
	}
	if t := param.ChildByFieldName("type"); t != nil {
		key = goTypeName(w.text(t)) + "." + name
	}
	assert.That(key == "" || strings.HasSuffix(key, "."+name), "the key ends with the method name")
	return variable, key
}

// goTypeName strips the pointer and the type arguments from a receiver type.
func goTypeName(t string) string {
	assert.That(t != "", "a receiver has a type")
	t = strings.TrimLeft(t, "* ")
	if cut := strings.IndexByte(t, '['); cut >= 0 {
		t = t[:cut]
	}
	assert.That(!strings.HasPrefix(t, "*") && !strings.Contains(t, "["), "the type name is bare")
	return t
}

// goLoopUnbounded reports whether a for statement has no condition, or the condition `true`.
func goLoopUnbounded(n *ts.Node, src []byte) bool {
	assert.That(int(n.EndByte()) <= len(src), "the node lies inside the source")
	assert.That(n.Kind() == "for_statement", "n is a for statement")
	for i := uint(0); i < n.NamedChildCount(); i++ {
		child := n.NamedChild(i)
		switch child.Kind() {
		case "range_clause":
			return false
		case "for_clause":
			return child.ChildByFieldName("condition") == nil
		case "block":
		default:
			return child.Utf8Text(src) == "true"
		}
	}
	return true
}

// goDiscard records an assignment whose last left operand is `_` and whose
// right side is one call: `_ = f()` or `x, _ := f()`. A conversion to a
// parenthesized type, as in `var _ I = (*T)(nil)`, is not a call.
func (w *walker) goDiscard(n *ts.Node) error {
	assert.That(w.lang == Go, "the file is Go")
	left, right := n.ChildByFieldName("left"), n.ChildByFieldName("right")
	names := 0
	var last *ts.Node
	if n.Kind() == "var_spec" {
		if n.ChildByFieldName("type") != nil {
			return nil // `var _ I = T(nil)` checks at compile time that T implements I
		}
		right = n.ChildByFieldName("value")
		for i := uint(0); i < n.NamedChildCount() && n.NamedChild(i).Kind() == "identifier"; i++ {
			last, names = n.NamedChild(i), names+1
		}
	} else if left != nil {
		last, names = lastNamedChild(left, "identifier"), countNamed(left, "identifier")
	}
	if last == nil || w.text(last) != "_" || right == nil || right.NamedChildCount() != 1 {
		return nil
	}
	call := right.NamedChild(0)
	if call.Kind() != "call_expression" {
		return nil
	}
	if fn := call.ChildByFieldName("function"); fn != nil && fn.Kind() == "parenthesized_expression" {
		return nil // a conversion `(*T)(nil)`, for example `var _ I = (*T)(nil)`
	}
	assert.That(names >= 1, "the left side has one or more names")
	detail := "discarded call result `_ = f(...)`"
	if names > 1 {
		detail = "discarded call result `x, _ := f(...)`"
	}
	return w.addSite(n, SiteDiscard, detail, 0)
}

// goGlobal records each name of a package-level var declaration.
func (w *walker) goGlobal(n *ts.Node) error {
	assert.That(n.Kind() == "var_declaration", "n is a var declaration")
	assert.That(w.lang == Go, "the file is Go")
	if p := n.Parent(); p == nil || p.Kind() != "source_file" {
		return nil
	}
	for i := uint(0); i < n.NamedChildCount(); i++ {
		spec := n.NamedChild(i)
		if spec.Kind() != "var_spec" {
			continue
		}
		if name := spec.ChildByFieldName("name"); name != nil {
			if err := w.addSite(spec, SiteGlobal, "package variable `"+w.text(name)+"`", 0); err != nil {
				return err
			}
		}
	}
	return nil
}

// goComment records a //go: directive comment.
func (w *walker) goComment(n *ts.Node) error {
	assert.That(n.Kind() == "comment", "n is a comment")
	assert.That(w.lang == Go, "the file is Go")
	t := w.text(n)
	if !strings.HasPrefix(t, "//go:") {
		return nil
	}
	name := strings.TrimPrefix(t, "//")
	if cut := strings.IndexAny(name, " \t"); cut >= 0 {
		name = name[:cut]
	}
	return w.addSite(n, SiteDirective, name, 0)
}

// lastNamedChild returns the last named child of n with the given kind, or nil.
func lastNamedChild(n *ts.Node, kind string) *ts.Node {
	assert.That(n != nil, "n is not nil")
	assert.That(kind != "", "the kind is not empty")
	var last *ts.Node
	for i := uint(0); i < n.NamedChildCount(); i++ {
		if c := n.NamedChild(i); c.Kind() == kind {
			last = c
		} else {
			last = nil
		}
	}
	assert.That(last == nil || last.Kind() == kind, "the result has the requested kind")
	return last
}

// countNamed counts the named children of n with the given kind.
func countNamed(n *ts.Node, kind string) int {
	assert.That(n != nil, "n is not nil")
	assert.That(kind != "", "the kind is not empty")
	count := 0
	for i := uint(0); i < n.NamedChildCount(); i++ {
		if n.NamedChild(i).Kind() == kind {
			count++
		}
	}
	assert.That(count <= int(n.NamedChildCount()), "the count is within the child count")
	return count
}

// pointerChain records a chain of nested pointer nodes once, at the outermost
// node. A pointer to a function counts as a function pointer, not as depth.
func (w *walker) pointerChain(n *ts.Node, outer, inner string) (bool, error) {
	assert.That(n.Kind() == outer || n.Kind() == inner, "n is a pointer node")
	if p := n.Parent(); p != nil && (p.Kind() == outer || p.Kind() == inner) {
		return true, nil
	}
	depth, cur := 0, n
	for steps := 0; cur != nil && steps < w.limits.MaxDepth; steps++ {
		k := cur.Kind()
		if k != outer && k != inner {
			break
		}
		depth++
		cur = firstTypeChild(cur)
	}
	if cur != nil && (cur.Kind() == "function_signature" || cur.Kind() == "function_type") {
		return true, w.addSite(n, SiteFuncPointer, "pointer to a function", 0)
	}
	assert.That(depth >= 1, "a pointer chain has one or more levels")
	return true, w.addSite(n, SitePointer, "", depth)
}

// firstTypeChild returns the "type" field of n, else the first named child.
func firstTypeChild(n *ts.Node) *ts.Node {
	assert.That(n != nil, "n is not nil")
	out := n.ChildByFieldName("type")
	if out == nil {
		out = n.ChildByFieldName("declarator")
	}
	if out == nil {
		out = n.NamedChild(0)
	}
	assert.That(out == nil || out.StartByte() >= n.StartByte(), "the child lies inside n")
	return out
}
