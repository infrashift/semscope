package engine

import (
	"fmt"
	"sort"

	"github.com/infrashift/semscope/internal/assert"
	"github.com/infrashift/semscope/internal/parser"
)

// assertCounts returns the assertion count of each function of f that the
// assertion check counts. A test function and an assertion primitive (a
// function whose own name matches an assertion pattern) have no entry.
func assertCounts(f *parser.File, patterns []string) map[int]int {
	assert.That(f != nil && !f.Exempt, "the callers skip exempt files")
	counts := map[int]int{}
	for i, fn := range f.Functions {
		self := fn.Name
		if f.Lang == parser.Go && f.Package != "" {
			self = f.Package + "." + fn.Name
		}
		primitive := matchCall(patterns, parser.Call{Callee: self, Name: fn.Name})
		if !fn.Exempt && !primitive {
			counts[i] = 0
		}
	}
	for _, call := range f.Calls {
		if _, counted := counts[call.Fn]; counted && matchCall(patterns, call) {
			counts[call.Fn]++
		}
	}
	assert.That(len(counts) <= len(f.Functions), "one count for each counted function")
	return counts
}

// assertionsPerFunction reports each function with fewer than min_assertions
// assertions, in the mode per_function.
func (c *fileCtx) assertionsPerFunction(cr *compiledRule) {
	assert.That(cr.rule.Check.Kind == "min_assertions", "the rule has the kind of this checker")
	p, ok := cr.rule.Params["mode"]
	if !ok || p.Str != "per_function" {
		return
	}
	limit := cr.intParam("min_assertions", 2)
	counts := assertCounts(c.f, cr.rule.Check.AssertCalls[string(c.f.Lang)])
	indexes := make([]int, 0, len(counts))
	for i := range counts {
		indexes = append(indexes, i)
	}
	sort.Ints(indexes)
	for _, i := range indexes {
		if fn := c.f.Functions[i]; counts[i] < limit {
			at := parser.Pos{Line: fn.Pos.Line, Col: fn.Pos.Col, EndLine: fn.Pos.Line, EndCol: fn.Pos.Col}
			c.add(cr, at, fmt.Sprintf("function `%s` has %d assertion(s) (min %d)", fn.Name, counts[i], limit))
		}
	}
	assert.That(limit >= 0, "the schema requires a non-negative min_assertions")
}

// assertionAverages computes the functions and assertions of each language
// and reports each language whose average is below min_assertions, in the
// mode average. The finding has no location.
func assertionAverages(cr *compiledRule, files []*parser.File) ([]Finding, map[parser.Language]LangStats) {
	assert.That(cr.rule.Check.Kind == "min_assertions", "the rule has the kind of this checker")
	stats := map[parser.Language]LangStats{}
	for _, f := range files {
		if f.Exempt || !cr.langs[f.Lang] {
			continue
		}
		s := stats[f.Lang]
		for _, n := range assertCounts(f, cr.rule.Check.AssertCalls[string(f.Lang)]) {
			s.Functions++
			s.Assertions += n
		}
		stats[f.Lang] = s
	}
	p, ok := cr.rule.Params["mode"]
	if ok && p.Str != "average" {
		return nil, stats
	}
	limit := cr.intParam("min_assertions", 2)
	var out []Finding
	for _, lang := range parser.Languages() {
		s := stats[lang]
		if s.Functions == 0 || float64(s.Assertions)/float64(s.Functions) >= float64(limit) {
			continue
		}
		out = append(out, Finding{RuleID: cr.rule.ID, Severity: cr.rule.Severity, Message: fmt.Sprintf(
			"%s: %d assertions in %d functions (average %.2f, min %d)",
			lang, s.Assertions, s.Functions, float64(s.Assertions)/float64(s.Functions), limit)})
	}
	assert.That(len(out) <= len(parser.Languages()), "one finding for each language at most")
	return out, stats
}
