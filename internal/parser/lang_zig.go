package parser

import (
	"strings"

	ts "github.com/tree-sitter/go-tree-sitter"

	"github.com/infrashift/semscope/internal/assert"
)

// visitZig records the Zig constructs of node n. The checks match
// scripts/nasa-lint.py; the wiki spec "nasa-lint.py check inventory" lists them.
func (w *walker) visitZig(n *ts.Node, kind string) (bool, error) {
	assert.That(kind == n.Kind(), "visit passes the kind of n")
	assert.That(w.lang == Zig, "the file is Zig")
	switch kind {
	case "test_declaration":
		return false, nil
	case "function_declaration":
		return true, w.zigFunction(n)
	case "call_expression":
		if fn := n.ChildByFieldName("function"); fn != nil {
			return true, w.addCall(n, w.text(fn))
		}
	case "builtin_function":
		if id := n.NamedChild(0); id != nil {
			return true, w.addSite(n, SiteDirective, w.text(id), 0)
		}
	case "while_statement":
		if cond := n.ChildByFieldName("condition"); cond != nil && w.text(cond) == "true" {
			return true, w.addSite(n, SiteUnboundedLoop, "unbounded loop `while (true)`", 0)
		}
	case "variable_declaration", "assignment_expression":
		return true, w.zigDeclaration(n)
	case "catch_expression":
		return true, w.zigCatch(n)
	case "else_clause":
		return true, w.zigElse(n)
	case "pointer_type":
		return w.pointerChain(n, "pointer_type", "pointer_type")
	}
	return true, nil
}

// zigFunction records a Zig function declaration with a body.
func (w *walker) zigFunction(n *ts.Node) error {
	name := n.ChildByFieldName("name")
	if name == nil || n.ChildByFieldName("body") == nil {
		return nil
	}
	fn := Function{Name: w.text(name)}
	fn.Exempt = strings.HasPrefix(fn.Name, "test") || strings.HasSuffix(fn.Name, "Test")
	assert.That(fn.Name != "", "a function has a name")
	assert.That(fn.Key == "", "addFunction sets the key")
	return w.addFunction(n, fn)
}

// zigDeclaration records an import, a container-level var, or a discarded
// call result `_ = f()` or `_ = try f()`.
func (w *walker) zigDeclaration(n *ts.Node) error {
	assert.That(n.Kind() == "variable_declaration" || n.Kind() == "assignment_expression", "n is a declaration or an assignment")
	assert.That(w.lang == Zig, "the file is Zig")
	target, value := n.NamedChild(0), n.NamedChild(n.NamedChildCount()-1)
	if n.Kind() == "assignment_expression" {
		target, value = n.ChildByFieldName("left"), n.ChildByFieldName("right")
	}
	if target == nil || value == nil || target.Equals(*value) {
		return nil
	}
	if value.Kind() == "builtin_function" && strings.HasPrefix(w.text(value), "@import(") {
		w.f.Imports[w.text(target)] = strings.Trim(strings.TrimPrefix(w.text(value), "@import"), "()\" ")
	}
	if n.Kind() == "variable_declaration" && w.enclosing() < 0 && zigIsVar(n, w.src) {
		if err := w.addSite(n, SiteGlobal, "container-level `var "+w.text(target)+"`", 0); err != nil {
			return err
		}
	}
	if w.text(target) != "_" {
		return nil
	}
	if value.Kind() == "try_expression" && value.NamedChildCount() > 0 {
		value = value.NamedChild(0)
	}
	if value.Kind() != "call_expression" {
		return nil
	}
	return w.addSite(n, SiteDiscard, "discarded call result `_ = f(...)`", 0)
}

// zigIsVar reports whether a variable declaration uses the keyword var.
func zigIsVar(n *ts.Node, src []byte) bool {
	assert.That(n.Kind() == "variable_declaration", "n is a variable declaration")
	assert.That(n.EndByte() <= uint(len(src)), "the node lies inside the source")
	for i := uint(0); i < n.ChildCount(); i++ {
		switch n.Child(i).Utf8Text(src) {
		case "var":
			return true
		case "const":
			return false
		}
	}
	return false
}

// zigCatch records `catch {}` and `catch unreachable`.
func (w *walker) zigCatch(n *ts.Node) error {
	assert.That(n.Kind() == "catch_expression", "n is a catch expression")
	assert.That(w.lang == Zig, "the file is Zig")
	last := n.Child(n.ChildCount() - 1)
	if last == nil || countNamed(n, "payload") > 0 {
		return nil
	}
	switch {
	case last.Kind() == "block" && last.NamedChildCount() == 0:
		return w.addSite(n, SiteDiscard, "swallowed error `catch {}`", 0)
	case last.Utf8Text(w.src) == "unreachable":
		return w.addSite(n, SiteDiscard, "`catch unreachable` hides an error", 0)
	}
	return nil
}

// zigElse records `else |_| {}`.
func (w *walker) zigElse(n *ts.Node) error {
	assert.That(n.Kind() == "else_clause", "n is an else clause")
	assert.That(w.lang == Zig, "the file is Zig")
	payload, body := n.NamedChild(0), n.NamedChild(n.NamedChildCount()-1)
	if payload == nil || body == nil || payload.Kind() != "payload" || w.text(payload) != "|_|" {
		return nil
	}
	for steps := 0; body.Kind() != "block" && body.NamedChildCount() > 0 && steps < 4; steps++ {
		body = body.NamedChild(0)
	}
	if body.Kind() == "block" && body.NamedChildCount() == 0 {
		return w.addSite(n, SiteDiscard, "swallowed error `else |_| {}`", 0)
	}
	return nil
}
