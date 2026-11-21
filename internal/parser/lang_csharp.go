package parser

import (
	"strings"

	ts "github.com/infrashift/ztree/bindings/go"

	"github.com/infrashift/semscope/internal/assert"
)

// visitCSharp records the C# constructs of node n.
func (w *walker) visitCSharp(n *ts.Node, kind string) (bool, error) {
	assert.That(kind == n.Kind(), "visit passes the kind of n")
	assert.That(w.lang == CSharp, "the file is C#")
	switch kind {
	case "class_declaration":
		return !csharpHasTestAttribute(n, w.src), nil
	case "method_declaration", "constructor_declaration", "destructor_declaration", "local_function_statement",
		"operator_declaration", "conversion_operator_declaration", "accessor_declaration":
		return true, w.csharpFunction(n)
	case "invocation_expression":
		if fn := n.ChildByFieldName("function"); fn != nil {
			return true, w.addCall(n, w.text(fn))
		}
	case "goto_statement":
		return true, w.addSite(n, SiteGoto, "`goto` statement", 0)
	case "while_statement", "do_statement":
		if cond := n.ChildByFieldName("condition"); cond != nil && w.text(cond) == "true" {
			return true, w.addSite(n, SiteUnboundedLoop, "unbounded loop `while (true)`", 0)
		}
	case "for_statement":
		if n.ChildByFieldName("condition") == nil {
			return true, w.addSite(n, SiteUnboundedLoop, "`for` loop without a condition", 0)
		}
	case "field_declaration":
		return true, w.csharpField(n)
	case "assignment_expression":
		return true, w.csharpDiscard(n)
	case "catch_clause":
		return true, w.csharpCatch(n)
	case "preproc_define", "preproc_undef", "preproc_pragma":
		return true, w.addSite(n, SiteDirective, strings.TrimPrefix(kind, "preproc_"), 0)
	case "preproc_if":
		return true, w.addSite(n, SiteConditional, kind, cConditionalDepth(n, w.limits.MaxDepth))
	case "pointer_type":
		return w.pointerChain(n, "pointer_type", "pointer_type")
	case "function_pointer_type":
		return true, w.addSite(n, SiteFuncPointer, "function pointer type", 0)
	}
	return true, nil
}

// csharpFunction records a C# method, constructor, destructor, operator,
// accessor or local function with a body. The key is "Type.name", with the
// innermost type that holds the function. A test method is exempt.
func (w *walker) csharpFunction(n *ts.Node) error {
	assert.That(w.lang == CSharp, "the file is C#")
	assert.That(n.Parent() != nil, "a function has a parent")
	if n.ChildByFieldName("body") == nil {
		return nil
	}
	name := w.csharpFunctionName(n)
	if name == "" {
		return nil
	}
	fn := Function{Name: name, Exempt: csharpHasTestAttribute(n, w.src)}
	if typ := w.csharpTypeName(n); typ != "" {
		fn.Key = typ + "." + name
	}
	return w.addFunction(n, fn)
}

// csharpFunctionName returns the name of a C# function node, or "". An
// operator gets the name "operator" and its symbol or type. An accessor gets
// the name of its property and its own name, for example "Count.get".
func (w *walker) csharpFunctionName(n *ts.Node) string {
	assert.That(n != nil, "n is not nil")
	assert.That(w.lang == CSharp, "the file is C#")
	switch n.Kind() {
	case "operator_declaration":
		if op := n.ChildByFieldName("operator"); op != nil {
			return "operator" + w.text(op)
		}
		return ""
	case "conversion_operator_declaration":
		if t := n.ChildByFieldName("type"); t != nil {
			return "operator " + w.text(t)
		}
		return ""
	case "accessor_declaration":
		return w.csharpAccessorName(n)
	}
	name := n.ChildByFieldName("name")
	if name == nil {
		return ""
	}
	if n.Kind() == "destructor_declaration" {
		return "~" + w.text(name)
	}
	return w.text(name)
}

// csharpAccessorName returns the name of an accessor: the name of its
// property or event and the accessor name. An indexer has the name "this".
func (w *walker) csharpAccessorName(n *ts.Node) string {
	assert.That(n.Kind() == "accessor_declaration", "n is an accessor")
	assert.That(w.lang == CSharp, "the file is C#")
	acc := n.ChildByFieldName("name")
	list := n.Parent()
	if acc == nil || list == nil || list.Parent() == nil {
		return ""
	}
	owner := "this"
	if name := list.Parent().ChildByFieldName("name"); name != nil {
		owner = w.text(name)
	}
	return owner + "." + w.text(acc)
}

// csharpTypeName returns the name of the innermost class, struct, record or
// interface that holds n, or "" for a function outside each type.
func (w *walker) csharpTypeName(n *ts.Node) string {
	assert.That(n != nil, "n is not nil")
	assert.That(w.limits.MaxDepth > 0, "New checks the limits")
	cur := n.Parent()
	for steps := 0; cur != nil && steps < w.limits.MaxDepth; steps++ {
		switch cur.Kind() {
		case "class_declaration", "struct_declaration", "record_declaration", "interface_declaration":
			if name := cur.ChildByFieldName("name"); name != nil {
				return w.text(name)
			}
			return ""
		}
		cur = cur.Parent()
	}
	return ""
}

// csharpField records a static field that is not readonly and not const.
func (w *walker) csharpField(n *ts.Node) error {
	assert.That(n.Kind() == "field_declaration", "n is a field declaration")
	assert.That(w.lang == CSharp, "the file is C#")
	static, fixed := false, false
	for i := uint(0); i < n.NamedChildCount(); i++ {
		c := n.NamedChild(i)
		if c.Kind() != "modifier" {
			continue
		}
		switch w.text(c) {
		case "static":
			static = true
		case "readonly", "const":
			fixed = true
		}
	}
	if !static || fixed {
		return nil
	}
	return w.addSite(n, SiteGlobal, "mutable `static` field", 0)
}

// csharpDiscard records the statement `_ = f();`.
func (w *walker) csharpDiscard(n *ts.Node) error {
	assert.That(n.Kind() == "assignment_expression", "n is an assignment")
	assert.That(w.lang == CSharp, "the file is C#")
	left, right := n.ChildByFieldName("left"), n.ChildByFieldName("right")
	if left == nil || right == nil || w.text(left) != "_" {
		return nil
	}
	if p := n.Parent(); p == nil || p.Kind() != "expression_statement" {
		return nil
	}
	if k := right.Kind(); k != "invocation_expression" && k != "await_expression" {
		return nil
	}
	return w.addSite(n, SiteDiscard, "discarded call result `_ = f()`", 0)
}

// csharpCatch records a catch clause with an empty body and no type, or the
// type Exception.
func (w *walker) csharpCatch(n *ts.Node) error {
	assert.That(n.Kind() == "catch_clause", "n is a catch clause")
	assert.That(w.lang == CSharp, "the file is C#")
	body := n.ChildByFieldName("body")
	if body == nil || body.NamedChildCount() != uint(countNamed(body, "comment")) {
		return nil
	}
	for i := uint(0); i < n.NamedChildCount(); i++ {
		c := n.NamedChild(i)
		if c.Kind() == "catch_filter_clause" {
			return nil
		}
		if t := c.ChildByFieldName("type"); c.Kind() == "catch_declaration" && t != nil {
			if name := w.text(t); name != "Exception" && name != "System.Exception" {
				return nil
			}
		}
	}
	return w.addSite(n, SiteDiscard, "swallowed exception, an empty `catch` block", 0)
}

// csharpHasTestAttribute reports whether n has an attribute of MSTest, NUnit
// or xUnit that marks a test class or a test method.
func csharpHasTestAttribute(n *ts.Node, src []byte) bool {
	assert.That(n != nil, "n is not nil")
	assert.That(src != nil, "the source exists")
	for i := uint(0); i < n.NamedChildCount(); i++ {
		list := n.NamedChild(i)
		if list.Kind() != "attribute_list" {
			continue
		}
		for j := uint(0); j < list.NamedChildCount(); j++ {
			if name := list.NamedChild(j).ChildByFieldName("name"); name != nil && isCSharpTestAttribute(name.Utf8Text(src)) {
				return true
			}
		}
	}
	return false
}

// isCSharpTestAttribute reports whether name, without its namespace and its
// Attribute suffix, is a test attribute.
func isCSharpTestAttribute(name string) bool {
	assert.That(name != "", "an attribute has a name")
	if cut := strings.LastIndex(name, "."); cut >= 0 {
		name = name[cut+1:]
	}
	name = strings.TrimSuffix(name, "Attribute")
	assert.That(!strings.Contains(name, "."), "the namespace is gone")
	switch name {
	case "TestClass", "TestFixture", "TestMethod", "Test", "TestCase", "Fact", "Theory":
		return true
	}
	return false
}
