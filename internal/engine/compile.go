package engine

import (
	"fmt"
	"math"
	"path"
	"regexp"
	"strconv"
	"strings"

	ts "github.com/infrashift/ztree/bindings/go"

	"github.com/infrashift/semscope/internal/assert"
	"github.com/infrashift/semscope/internal/parser"
	"github.com/infrashift/semscope/internal/ruleset"
)

// violationCapture is the capture name that marks a finding in an ast_query rule.
const violationCapture = "violation"

// compiledRule is one active rule with its patterns and queries compiled.
type compiledRule struct {
	rule ruleset.Rule
	// tag is the nasa-lint suppression tag for a NASA-P10 rule, for example "R7", else "".
	tag     string
	langs   map[parser.Language]bool
	bans    map[parser.Language][]*regexp.Regexp
	queries map[parser.Language]*ts.Query
	capture map[parser.Language]uint32
}

// compileRule compiles the token_ban patterns and the ast_query queries of r.
func compileRule(r ruleset.Rule) (compiledRule, error) {
	cr := compiledRule{
		rule: r, tag: nasaTag(r.ID), langs: map[parser.Language]bool{},
		bans: map[parser.Language][]*regexp.Regexp{}, queries: map[parser.Language]*ts.Query{},
		capture: map[parser.Language]uint32{},
	}
	for _, l := range r.Languages {
		cr.langs[parser.Language(l)] = true
	}
	for lang, patterns := range r.Check.Patterns {
		for _, p := range patterns {
			re, err := regexp.Compile(p)
			if err != nil {
				return compiledRule{}, fmt.Errorf("engine: rule %s: %s pattern %q: %w", r.ID, lang, p, err)
			}
			cr.bans[parser.Language(lang)] = append(cr.bans[parser.Language(lang)], re)
		}
	}
	for lang, src := range r.Check.Query {
		if err := cr.compileQuery(parser.Language(lang), src); err != nil {
			cr.close()
			return compiledRule{}, err
		}
	}
	assert.That(len(cr.langs) == len(r.Languages), "each language of the rule is known")
	assert.That(len(cr.queries) == len(cr.capture), "each query has a violation capture")
	return cr, nil
}

// compileQuery compiles one Tree-sitter query. The query needs a @violation capture.
func (cr *compiledRule) compileQuery(lang parser.Language, src string) error {
	assert.That(cr.queries != nil && cr.capture != nil, "compileRule makes the query maps")
	grammar, ok := parser.Grammar(lang)
	if !ok {
		return fmt.Errorf("engine: rule %s: unknown query language %q", cr.rule.ID, lang)
	}
	q, qerr := ts.NewQuery(grammar, src)
	if qerr != nil {
		return fmt.Errorf("engine: rule %s: %s query: %s", cr.rule.ID, lang, qerr.Error())
	}
	idx, found := q.CaptureIndexForName(violationCapture)
	if !found {
		return fmt.Errorf("engine: rule %s: %s query has no @%s capture", cr.rule.ID, lang, violationCapture)
	}
	assert.That(idx < uint(len(q.CaptureNames())), "the capture index is valid")
	if idx > math.MaxUint32 {
		return fmt.Errorf("engine: rule %s: %s query has too many captures", cr.rule.ID, lang)
	}
	cr.queries[lang], cr.capture[lang] = q, uint32(idx)
	assert.That(cr.queries[lang] != nil, "the query is stored")
	return nil
}

func (cr *compiledRule) close() {
	assert.That(cr.queries != nil, "the rule has a query map")
	for lang := range cr.queries {
		delete(cr.queries, lang)
	}
	assert.That(len(cr.queries) == 0, "each query is closed")
}

// nasaTag returns "R<n>" for the rule ID NASA-P10-<nn>, else "".
func nasaTag(id string) string {
	num, ok := strings.CutPrefix(id, "NASA-P10-")
	n, err := strconv.Atoi(num)
	if !ok || err != nil || n < 1 || n > 10 {
		return ""
	}
	tag := "R" + strconv.Itoa(n)
	assert.That(len(tag) == 2 || tag == "R10", "a tag is R1 to R10")
	assert.That(strings.HasPrefix(id, "NASA-P10-"), "only a NASA-P10 rule has a tag")
	return tag
}

// matchCall reports whether a call matches one of the patterns. A pattern
// with "." or ":" matches the whole callee text. Another pattern matches the
// last segment. A pattern can hold the wildcards of path.Match.
func matchCall(patterns []string, c parser.Call) bool {
	assert.That(strings.HasSuffix(c.Callee, c.Name), "the name is the last segment of the callee")
	for _, p := range patterns {
		subject := c.Name
		if strings.ContainsAny(p, ".:") {
			subject = c.Callee
		}
		if ok, err := path.Match(p, subject); err == nil && ok {
			return true
		}
	}
	return false
}

// matchName reports whether name matches one of the path.Match patterns.
func matchName(patterns []string, name string) bool {
	return matchCall(patterns, parser.Call{Callee: name, Name: name})
}

// intParam returns the integer parameter name of r, or fallback when r has no such parameter.
func (cr *compiledRule) intParam(name string, fallback int) int {
	assert.That(name != "", "the parameter has a name")
	assert.That(cr.rule.Params != nil, "decodeRule sets the params map")
	p, ok := cr.rule.Params[name]
	if !ok || p.Kind != ruleset.ParamInt {
		return fallback
	}
	return int(p.Int)
}

// boolParam returns the boolean parameter name of r, or false.
func (cr *compiledRule) boolParam(name string) bool {
	assert.That(name != "", "the parameter has a name")
	assert.That(cr.rule.Params != nil, "decodeRule sets the params map")
	p, ok := cr.rule.Params[name]
	return ok && p.Kind == ruleset.ParamBool && p.Bool
}

// suppressed reports whether line (1-based) or the line above holds the tag
// `semscope: allow(<ID>)` or, for a NASA-P10 rule, `nasa-lint: allow(R<n>)`.
func (cr *compiledRule) suppressed(lines []string, line int) bool {
	assert.That(line >= 1, "lines are 1-based")
	assert.That(cr.rule.ID != "", "a rule has an ID")
	for _, i := range []int{line - 1, line - 2} {
		if i < 0 || i >= len(lines) {
			continue
		}
		if strings.Contains(lines[i], "semscope: allow("+cr.rule.ID+")") {
			return true
		}
		if cr.tag != "" && strings.Contains(lines[i], "nasa-lint: allow("+cr.tag+")") {
			return true
		}
	}
	return false
}
