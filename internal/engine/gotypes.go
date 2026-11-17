package engine

import (
	"path/filepath"
	"strings"

	"github.com/infrashift/semscope/internal/assert"
	"github.com/infrashift/semscope/internal/gotypes"
	"github.com/infrashift/semscope/internal/parser"
)

// goTypeChecks runs the type-aware Go check of each active checked_returns
// rule with go_types set to true on the Go files of a scan. files[i] is the parsed record of sources[i]. A test
// file is exempt.
func (e *Engine) goTypeChecks(sources []source, files []*parser.File) ([]Finding, error) {
	assert.That(len(sources) == len(files), "one record for each source")
	var rules []*compiledRule
	for i := range e.rules {
		cr := &e.rules[i]
		if cr.rule.Check.Kind == "checked_returns" && cr.langs[parser.Go] && cr.boolParam("go_types") {
			rules = append(rules, cr)
		}
	}
	if len(rules) == 0 {
		return nil, nil
	}
	byAbs := map[string]*parser.File{}
	var paths []string
	for i, f := range files {
		if f.Lang != parser.Go || f.Exempt || strings.HasSuffix(f.Path, "_test.go") {
			continue
		}
		byAbs[sources[i].abs] = f
		paths = append(paths, sources[i].abs)
	}
	if len(paths) == 0 {
		return nil, nil
	}
	drops, err := gotypes.Find(paths, e.limits.MaxFiles)
	if err != nil {
		return nil, err
	}
	var out []Finding
	for _, d := range drops {
		f := byAbs[d.File]
		assert.That(f != nil, "Find reports the given files only")
		for _, cr := range rules {
			if d.Ignored(listParam(cr, "go_ignore_calls")) || cr.suppressed(f.Lines, d.Line) {
				continue
			}
			out = append(out, Finding{
				RuleID: cr.rule.ID, Severity: cr.rule.Severity, Message: dropMessage(d), Path: f.Path,
				Line: d.Line, Col: d.Col, EndLine: d.EndLine, EndCol: d.EndCol,
			})
		}
	}
	assert.That(filepath.IsAbs(paths[0]), "collect makes the paths absolute")
	return out, nil
}

// listParam returns the list parameter name of cr, or nil.
func listParam(cr *compiledRule, name string) []string {
	assert.That(name != "", "the parameter has a name")
	assert.That(cr.rule.Params != nil, "decodeRule sets the params map")
	p, ok := cr.rule.Params[name]
	if !ok {
		return nil
	}
	return p.List
}

// dropMessage describes one drop.
func dropMessage(d gotypes.Drop) string {
	assert.That(d.Callee != "", "a drop names its callee")
	msg := "unchecked error of `" + d.Callee + "`"
	switch d.Stmt {
	case gotypes.StmtDefer:
		msg += " in a defer statement"
	case gotypes.StmtGo:
		msg += " in a go statement"
	}
	assert.That(d.Stmt != "", "a drop has a statement kind")
	return msg
}
