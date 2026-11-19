package parser

import (
	ts "github.com/infrashift/ztree/bindings/go"

	"github.com/infrashift/semscope/internal/assert"
)

// visitPython records the Python constructs of node n.
func (w *walker) visitPython(n *ts.Node, kind string) (bool, error) {
	assert.That(kind == n.Kind(), "visit passes the kind of n")
	assert.That(w.lang == Python, "the file is Python")
	switch kind {
	case "function_definition":
		name := n.ChildByFieldName("name")
		if name == nil {
			return true, nil
		}
		fn := Function{Name: w.text(name)}
		fn.Exempt = len(fn.Name) >= 4 && fn.Name[:4] == "test"
		return true, w.addFunction(n, fn)
	case "call":
		if fn := n.ChildByFieldName("function"); fn != nil {
			return true, w.addCall(n, w.text(fn))
		}
	case "assert_statement":
		return true, w.addCall(n, "assert")
	case "while_statement":
		if cond := n.ChildByFieldName("condition"); cond != nil && (w.text(cond) == "True" || w.text(cond) == "1") {
			return true, w.addSite(n, SiteUnboundedLoop, "unbounded loop `while True`", 0)
		}
	case "global_statement":
		return true, w.addSite(n, SiteGlobal, "`global` statement", 0)
	case "except_clause":
		return true, w.pythonExcept(n)
	}
	return true, nil
}

// pythonExcept records an except clause whose body is only `pass` and whose
// type is absent, Exception or BaseException.
func (w *walker) pythonExcept(n *ts.Node) error {
	var body *ts.Node
	broad := true
	for i := uint(0); i < n.NamedChildCount(); i++ {
		c := n.NamedChild(i)
		switch {
		case c.Kind() == "block":
			body = c
		case c.Kind() != "comment":
			t := w.text(c)
			broad = t == "Exception" || t == "BaseException"
		}
	}
	if body == nil || !broad || body.NamedChildCount() != 1 || body.NamedChild(0).Kind() != "pass_statement" {
		return nil
	}
	assert.That(body.Kind() == "block", "the except body is a block")
	assert.That(n.Kind() == "except_clause", "n is an except clause")
	return w.addSite(n, SiteDiscard, "swallowed exception `except: pass`", 0)
}
