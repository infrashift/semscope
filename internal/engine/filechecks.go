package engine

import (
	"fmt"
	"strings"

	ts "github.com/tree-sitter/go-tree-sitter"

	"github.com/infrashift/semscope/internal/assert"
	"github.com/infrashift/semscope/internal/parser"
)

// fileCtx holds the state of the per-file checks of one file.
type fileCtx struct {
	f        *parser.File
	src      []byte
	tree     *ts.Tree
	cursor   *ts.QueryCursor
	findings []Finding
}

// add records a finding of rule cr at p, unless a comment suppresses it.
func (c *fileCtx) add(cr *compiledRule, p parser.Pos, detail string) {
	assert.That(p.Line >= 1, "a finding has a 1-based line")
	assert.That(c.f.Path != "", "the file has a path")
	if cr.suppressed(c.f.Lines, p.Line) {
		return
	}
	c.findings = append(c.findings, Finding{
		RuleID: cr.rule.ID, Severity: cr.rule.Severity, Message: detail, Path: c.f.Path,
		Line: p.Line, Col: p.Col, EndLine: p.EndLine, EndCol: p.EndCol,
	})
}

// checkFile runs the per-file checks of each active rule on f. A switch
// selects the checker of each kind.
func (e *Engine) checkFile(f *parser.File, src []byte, tree *ts.Tree, cursor *ts.QueryCursor) []Finding {
	assert.That(tree != nil && cursor != nil, "the file has a tree and a query cursor")
	c := &fileCtx{f: f, src: src, tree: tree, cursor: cursor}
	for i := range e.rules {
		cr := &e.rules[i]
		if !cr.langs[f.Lang] {
			continue
		}
		lang := string(f.Lang)
		switch cr.rule.Check.Kind {
		case "no_goto_or_recursion":
			c.sites(cr, parser.SiteGoto)
			c.calls(cr, cr.rule.Check.BannedCalls[lang], false, "banned call")
		case "bounded_loops":
			c.sites(cr, parser.SiteUnboundedLoop)
		case "no_dynamic_alloc":
			c.alloc(cr)
		case "max_function_lines":
			c.length(cr)
		case "min_assertions":
			c.assertionsPerFunction(cr)
		case "narrowest_scope":
			c.sites(cr, parser.SiteGlobal)
		case "checked_returns":
			c.sites(cr, parser.SiteDiscard)
			c.calls(cr, cr.rule.Check.MustCheckCalls[lang], true, "unused result of")
		case "restricted_preprocessor":
			c.preprocessor(cr)
		case "max_pointer_indirection":
			c.pointers(cr)
		case "ast_query":
			c.query(cr)
		case "token_ban":
			c.bans(cr)
		}
	}
	assert.That(len(c.findings) == 0 || c.findings[0].Path == f.Path, "the findings belong to f")
	return c.findings
}

// sites reports each site of kind k.
func (c *fileCtx) sites(cr *compiledRule, k parser.SiteKind) {
	assert.That(k >= parser.SiteGoto && k <= parser.SiteFuncPointer, "the site kind is known")
	assert.That(cr != nil && c.f != nil, "the rule and the file exist")
	for _, s := range c.f.Sites {
		if s.Kind == k {
			c.add(cr, s.Pos, s.Detail)
		}
	}
}

// calls reports each call that matches patterns. With unusedOnly, it
// reports only a call whose result the code does not use.
func (c *fileCtx) calls(cr *compiledRule, patterns []string, unusedOnly bool, what string) {
	assert.That(cr != nil && c.f != nil, "the rule and the file exist")
	assert.That(what != "", "the finding names the call")
	if len(patterns) == 0 {
		return
	}
	for _, call := range c.f.Calls {
		if (!unusedOnly || call.Unused) && matchCall(patterns, call) {
			c.add(cr, call.Pos, what+" `"+call.Callee+"`")
		}
	}
}

// alloc reports each allocator call outside the init functions. A call at
// file scope runs at initialization, so it is not a finding.
func (c *fileCtx) alloc(cr *compiledRule) {
	assert.That(cr.rule.Check.Kind == "no_dynamic_alloc", "the rule has the kind of this checker")
	patterns := cr.rule.Check.AllocCalls[string(c.f.Lang)]
	for _, call := range c.f.Calls {
		if call.Fn < 0 || !matchCall(patterns, call) {
			continue
		}
		assert.That(call.Fn < len(c.f.Functions), "the call has a valid function index")
		if fn := c.f.Functions[call.Fn]; !matchName(cr.rule.Check.InitFunctions, fn.Name) {
			c.add(cr, call.Pos, fmt.Sprintf("allocation `%s` in function `%s`", call.Callee, fn.Name))
		}
	}
}

// length reports each function with more than max_lines lines.
func (c *fileCtx) length(cr *compiledRule) {
	assert.That(cr.rule.Check.Kind == "max_function_lines", "the rule has the kind of this checker")
	limit := cr.intParam("max_lines", 60)
	assert.That(limit > 0, "the schema requires a positive max_lines")
	for _, fn := range c.f.Functions {
		if n := fn.Pos.EndLine - fn.Pos.Line + 1; !fn.Exempt && n > limit {
			at := parser.Pos{Line: fn.Pos.Line, Col: fn.Pos.Col, EndLine: fn.Pos.Line, EndCol: fn.Pos.Col}
			c.add(cr, at, fmt.Sprintf("function `%s` has %d lines (max %d)", fn.Name, n, limit))
		}
	}
}

// preprocessor reports banned directives, banned dynamic-code calls and deep conditionals.
func (c *fileCtx) preprocessor(cr *compiledRule) {
	assert.That(cr.rule.Check.Kind == "restricted_preprocessor", "the rule has the kind of this checker")
	banned := cr.rule.Check.BannedDirectives[string(c.f.Lang)]
	limit := cr.intParam("max_conditional_depth", 2)
	assert.That(limit >= 0, "the schema requires a non-negative depth")
	for _, s := range c.f.Sites {
		switch {
		case s.Kind == parser.SiteDirective && matchName(banned, s.Detail):
			c.add(cr, s.Pos, "directive `"+s.Detail+"`")
		case s.Kind == parser.SiteConditional && s.Depth > limit:
			c.add(cr, s.Pos, fmt.Sprintf("conditional compilation nests %d levels (max %d)", s.Depth, limit))
		}
	}
	c.calls(cr, banned, false, "dynamic code")
}

// pointers reports deep pointer chains and, unless allowed, function pointers.
func (c *fileCtx) pointers(cr *compiledRule) {
	assert.That(cr.rule.Check.Kind == "max_pointer_indirection", "the rule has the kind of this checker")
	limit := cr.intParam("max_pointer_depth", 1)
	allowFn := cr.boolParam("allow_function_pointers")
	assert.That(limit >= 0, "the schema requires a non-negative depth")
	for _, s := range c.f.Sites {
		switch {
		case s.Kind == parser.SitePointer && s.Depth > limit:
			c.add(cr, s.Pos, fmt.Sprintf("%d levels of pointer indirection (max %d)", s.Depth, limit))
		case s.Kind == parser.SiteFuncPointer && !allowFn:
			c.add(cr, s.Pos, s.Detail)
		}
	}
}

// query reports each @violation capture of the ast_query query.
func (c *fileCtx) query(cr *compiledRule) {
	assert.That(cr.rule.Check.Kind == "ast_query", "the rule has the kind of this checker")
	assert.That(c.tree != nil && c.cursor != nil, "the file has a tree and a cursor")
	q, ok := cr.queries[c.f.Lang]
	if !ok {
		return
	}
	idx := cr.capture[c.f.Lang]
	matches := c.cursor.Matches(q, c.tree.RootNode(), c.src)
	for m := matches.Next(); m != nil; m = matches.Next() {
		for _, capture := range m.Captures {
			if capture.Index != idx {
				continue
			}
			text := strings.TrimSpace(capture.Node.Utf8Text(c.src))
			if len(text) > 40 {
				text = text[:40] + "..."
			}
			c.add(cr, parser.NodePos(&capture.Node), "construct `"+text+"`")
		}
	}
}

// bans reports each match of a token_ban pattern in the stripped source.
func (c *fileCtx) bans(cr *compiledRule) {
	assert.That(cr.rule.Check.Kind == "token_ban", "the rule has the kind of this checker")
	assert.That(len(c.f.Stripped) == len(c.f.Lines), "the stripped text keeps every line")
	for _, re := range cr.bans[c.f.Lang] {
		for i, line := range c.f.Stripped {
			for _, loc := range re.FindAllStringIndex(line, -1) {
				p := parser.Pos{Line: i + 1, Col: loc[0] + 1, EndLine: i + 1, EndCol: loc[1] + 1}
				c.add(cr, p, "banned pattern `"+re.String()+"`")
			}
		}
	}
}
