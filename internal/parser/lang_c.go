package parser

import (
	"regexp"
	"strings"

	ts "github.com/tree-sitter/go-tree-sitter"

	"github.com/infrashift/semscope/internal/assert"
)

// stringizeRE finds the # operator of a macro body, but not ##.
// semscope: allow(NASA-P10-06) the program compiles the pattern once and never writes it
var stringizeRE = regexp.MustCompile(`(^|[^#])#\s*[A-Za-z_]`)

// visitC records the C and C++ constructs of node n.
func (w *walker) visitC(n *ts.Node, kind string) (bool, error) {
	assert.That(kind == n.Kind(), "visit passes the kind of n")
	assert.That(w.lang == C || w.lang == CPP, "the file is C or C++")
	switch kind {
	case "function_definition":
		return true, w.cFunction(n)
	case "call_expression":
		if fn := n.ChildByFieldName("function"); fn != nil {
			return true, w.addCall(n, w.text(fn))
		}
	case "new_expression":
		return true, w.addCall(n, "new")
	case "goto_statement":
		return true, w.addSite(n, SiteGoto, "`goto` statement", 0)
	case "while_statement", "do_statement":
		if cond := n.ChildByFieldName("condition"); cond != nil && w.cConstantTrue(cond) {
			return true, w.addSite(n, SiteUnboundedLoop, "loop with a constant true condition", 0)
		}
	case "for_statement":
		if n.ChildByFieldName("condition") == nil {
			return true, w.addSite(n, SiteUnboundedLoop, "`for` loop without a condition", 0)
		}
	case "declaration":
		return true, w.cGlobal(n)
	case "preproc_function_def":
		return true, w.cMacro(n)
	case "preproc_call":
		if d := n.ChildByFieldName("directive"); d != nil {
			return true, w.addSite(n, SiteDirective, strings.TrimPrefix(w.text(d), "#"), 0)
		}
	case "preproc_if", "preproc_ifdef":
		return true, w.addSite(n, SiteConditional, kind, cConditionalDepth(n, w.limits.MaxDepth))
	case "pointer_declarator", "abstract_pointer_declarator":
		return w.pointerChain(n, "pointer_declarator", "abstract_pointer_declarator")
	case "function_declarator", "abstract_function_declarator":
		if d := n.ChildByFieldName("declarator"); d != nil && d.Kind() == "parenthesized_declarator" {
			return true, w.addSite(n, SiteFuncPointer, "function pointer", 0)
		}
	}
	return true, nil
}

// cFunction records a C or C++ function definition. The name is the
// innermost identifier of the declarator chain.
func (w *walker) cFunction(n *ts.Node) error {
	d := n.ChildByFieldName("declarator")
	for steps := 0; d != nil && d.Kind() != "function_declarator" && steps < w.limits.MaxDepth; steps++ {
		d = d.ChildByFieldName("declarator")
	}
	if d == nil {
		return nil
	}
	name := d.ChildByFieldName("declarator")
	if name == nil {
		return nil
	}
	if q := name.ChildByFieldName("name"); q != nil && name.Kind() == "qualified_identifier" {
		name = q
	}
	fn := Function{Name: w.text(name)}
	assert.That(fn.Name != "", "a function has a name")
	assert.That(d.Kind() == "function_declarator", "the chain ends at a function declarator")
	return w.addFunction(n, fn)
}

// cConstantTrue reports whether a parenthesized condition is the constant 1 or true.
func (w *walker) cConstantTrue(cond *ts.Node) bool {
	assert.That(w.lang == C || w.lang == CPP, "the file is C or C++")
	assert.That(cond != nil, "the caller checks the condition")
	c := strings.TrimSpace(strings.Trim(strings.TrimSpace(w.text(cond)), "()"))
	return c == "1" || c == "true"
}

// cGlobal records a file-scope variable that is not const and not a function prototype.
func (w *walker) cGlobal(n *ts.Node) error {
	assert.That(n.Kind() == "declaration", "n is a declaration")
	assert.That(w.lang == C || w.lang == CPP, "the file is C or C++")
	if p := n.Parent(); p == nil || p.Kind() != "translation_unit" {
		return nil
	}
	for i := uint(0); i < n.NamedChildCount(); i++ {
		c := n.NamedChild(i)
		if c.Kind() == "type_qualifier" && w.text(c) == "const" {
			return nil
		}
	}
	d := n.ChildByFieldName("declarator")
	for steps := 0; d != nil && steps < w.limits.MaxDepth; steps++ {
		switch d.Kind() {
		case "function_declarator":
			if inner := d.ChildByFieldName("declarator"); inner == nil || inner.Kind() != "parenthesized_declarator" {
				return nil
			}
			return w.addSite(n, SiteGlobal, "file-scope variable", 0)
		case "identifier":
			return w.addSite(n, SiteGlobal, "file-scope variable `"+w.text(d)+"`", 0)
		}
		d = d.ChildByFieldName("declarator")
	}
	return nil
}

// cMacro records the token pasting, stringizing and variadic forms of a function-like macro.
func (w *walker) cMacro(n *ts.Node) error {
	assert.That(n.Kind() == "preproc_function_def", "n is a macro definition")
	assert.That(w.lang == C || w.lang == CPP, "the file is C or C++")
	if params := n.ChildByFieldName("parameters"); params != nil && strings.Contains(w.text(params), "...") {
		if err := w.addSite(n, SiteDirective, "variadic_macro", 0); err != nil {
			return err
		}
	}
	value := n.ChildByFieldName("value")
	if value == nil {
		return nil
	}
	body := w.text(value)
	if strings.Contains(body, "##") {
		if err := w.addSite(n, SiteDirective, "token_paste", 0); err != nil {
			return err
		}
	}
	if stringizeRE.MatchString(body) {
		return w.addSite(n, SiteDirective, "stringize", 0)
	}
	return nil
}

// cConditionalDepth counts the #if and #ifdef nodes from n up to the root.
// maxDepth bounds the walk, because the tree depth is at most maxDepth.
func cConditionalDepth(n *ts.Node, maxDepth int) int {
	assert.That(maxDepth > 0, "New checks the limits")
	depth := 0
	cur := n
	for steps := 0; cur != nil && steps <= maxDepth; steps++ {
		if k := cur.Kind(); k == "preproc_if" || k == "preproc_ifdef" {
			depth++
		}
		cur = cur.Parent()
	}
	assert.That(depth >= 1, "n is a conditional")
	return depth
}
