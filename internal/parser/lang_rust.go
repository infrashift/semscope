package parser

import (
	"strings"

	ts "github.com/tree-sitter/go-tree-sitter"

	"github.com/infrashift/semscope/internal/assert"
)

// visitRust records the Rust constructs of node n.
func (w *walker) visitRust(n *ts.Node, kind string) (bool, error) {
	assert.That(kind == n.Kind(), "visit passes the kind of n")
	assert.That(w.lang == Rust, "the file is Rust")
	switch kind {
	case "mod_item":
		return !rustHasAttribute(n, w.src, "#[cfg(test)]"), nil
	case "function_item":
		return true, w.rustFunction(n)
	case "call_expression":
		return true, w.rustCall(n)
	case "macro_invocation":
		if m := n.ChildByFieldName("macro"); m != nil {
			return true, w.addCall(n, w.text(m)+"!")
		}
	case "macro_definition":
		return true, w.addSite(n, SiteDirective, "macro_rules!", 0)
	case "loop_expression":
		return true, w.addSite(n, SiteUnboundedLoop, "`loop` without a bound", 0)
	case "while_expression":
		if cond := n.ChildByFieldName("condition"); cond != nil && w.text(cond) == "true" {
			return true, w.addSite(n, SiteUnboundedLoop, "unbounded loop `while true`", 0)
		}
	case "let_declaration":
		return true, w.rustLet(n)
	case "static_item":
		if countNamed(n, "mutable_specifier") > 0 {
			return true, w.addSite(n, SiteGlobal, "`static mut` item", 0)
		}
	case "pointer_type", "reference_type":
		return w.pointerChain(n, "pointer_type", "reference_type")
	case "function_type":
		return true, w.addSite(n, SiteFuncPointer, "function pointer type", 0)
	}
	return true, nil
}

// rustFunction records a Rust function with a body. A #[test] function is exempt.
func (w *walker) rustFunction(n *ts.Node) error {
	name := n.ChildByFieldName("name")
	if name == nil || n.ChildByFieldName("body") == nil {
		return nil
	}
	fn := Function{Name: w.text(name), Exempt: rustHasAttribute(n, w.src, "#[test]")}
	assert.That(fn.Name != "", "a function has a name")
	assert.That(n.Kind() == "function_item", "n is a function item")
	return w.addFunction(n, fn)
}

// rustCall records a call. A call of unwrap or expect also counts as a
// hidden error, the same as `catch unreachable` in Zig.
func (w *walker) rustCall(n *ts.Node) error {
	assert.That(n.Kind() == "call_expression", "n is a call")
	assert.That(w.lang == Rust, "the file is Rust")
	fn := n.ChildByFieldName("function")
	if fn == nil {
		return nil
	}
	if err := w.addCall(n, w.text(fn)); err != nil {
		return err
	}
	if fn.Kind() == "field_expression" {
		if f := fn.ChildByFieldName("field"); f != nil && (w.text(f) == "unwrap" || w.text(f) == "expect") {
			return w.addSite(n, SiteDiscard, "`"+w.text(f)+"` hides an error", 0)
		}
	}
	return nil
}

// rustLet records `let _ = f();`.
func (w *walker) rustLet(n *ts.Node) error {
	assert.That(n.Kind() == "let_declaration", "n is a let declaration")
	assert.That(w.lang == Rust, "the file is Rust")
	pattern, value := n.ChildByFieldName("pattern"), n.ChildByFieldName("value")
	if pattern == nil || value == nil || w.text(pattern) != "_" {
		return nil
	}
	if k := value.Kind(); k != "call_expression" && k != "macro_invocation" && k != "try_expression" {
		return nil
	}
	return w.addSite(n, SiteDiscard, "discarded call result `let _ = f()`", 0)
}

// rustHasAttribute reports whether an attribute item with the text attr
// comes right before n.
func rustHasAttribute(n *ts.Node, src []byte, attr string) bool {
	assert.That(n != nil, "n is not nil")
	assert.That(strings.HasPrefix(attr, "#["), "attr is an attribute")
	prev := n.PrevNamedSibling()
	for steps := 0; prev != nil && steps < 16; steps++ {
		if prev.Kind() != "attribute_item" {
			return false
		}
		if strings.ReplaceAll(prev.Utf8Text(src), " ", "") == attr {
			return true
		}
		prev = prev.PrevNamedSibling()
	}
	return false
}
