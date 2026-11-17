package ruleset

import (
	"fmt"
	"path"
	"sort"

	"cuelang.org/go/cue"

	"github.com/infrashift/semscope/internal/assert"
)

// decodeRules decodes the validated rules map of v, sorted by ID.
func decodeRules(v cue.Value) ([]Rule, error) {
	assert.That(v.Exists(), "v exists")
	iter, err := v.LookupPath(cue.ParsePath("rules")).Fields()
	if err != nil {
		return nil, cueError("rules", err)
	}
	rules := []Rule{}
	for iter.Next() {
		if len(rules) >= MaxRules {
			return nil, fmt.Errorf("ruleset: more than %d rules", MaxRules)
		}
		rule, err := decodeRule(iter.Value())
		if err != nil {
			return nil, err
		}
		assert.That(rule.ID == iter.Selector().Unquoted(), "the schema sets id from the map key")
		rules = append(rules, rule)
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })
	return rules, nil
}

// decodeExcludes decodes the exclude map of v. A bad path pattern gives an error.
func decodeExcludes(v cue.Value) (map[string]string, error) {
	out := map[string]string{}
	ex := v.LookupPath(cue.ParsePath("exclude"))
	if !ex.Exists() {
		return out, nil
	}
	if err := ex.Decode(&out); err != nil {
		return nil, cueError("exclude", err)
	}
	if len(out) > MaxRules {
		return nil, fmt.Errorf("ruleset: more than %d exclude patterns", MaxRules)
	}
	for pattern := range out {
		if _, err := path.Match(pattern, ""); err != nil {
			return nil, fmt.Errorf("ruleset: exclude pattern %q: %w", pattern, err)
		}
	}
	assert.That(ex.Err() == nil, "the validated exclude map has no error")
	return out, nil
}

// addFlagExcludes adds each command-line pattern to excludes. A pattern
// that a ruleset also excludes keeps the reason of the ruleset.
func addFlagExcludes(excludes map[string]string, patterns []string) error {
	assert.That(excludes != nil, "decodeExcludes gives a map")
	for _, pattern := range patterns {
		if _, err := path.Match(pattern, ""); err != nil || pattern == "" {
			return fmt.Errorf("ruleset: --exclude pattern %q: want a path.Match pattern", pattern)
		}
		if _, ok := excludes[pattern]; !ok {
			excludes[pattern] = FlagExcludeReason
		}
	}
	if len(excludes) > MaxRules {
		return fmt.Errorf("ruleset: more than %d exclude patterns", MaxRules)
	}
	assert.That(len(patterns) == 0 || len(excludes) > 0, "the map holds the patterns")
	return nil
}

// optionalString returns the string at path in v, or "" when the field is absent.
func optionalString(v cue.Value, path string) string {
	assert.That(v.Exists(), "v exists")
	f := v.LookupPath(cue.ParsePath(path))
	if !f.Exists() {
		return ""
	}
	s, err := f.String()
	assert.That(err == nil, "the schema types the field as a string")
	return s
}

// decodeRule decodes one concrete rule.
func decodeRule(v cue.Value) (Rule, error) {
	var rule Rule
	if err := v.Decode(&rule); err != nil {
		return Rule{}, cueError("decode rule", err)
	}
	if !v.LookupPath(cue.ParsePath("enabled")).Exists() {
		rule.Enabled = true
	}
	if rule.Severity == "" {
		rule.Severity = "error"
	}
	params, err := decodeParams(v.LookupPath(cue.ParsePath("params")))
	if err != nil {
		return Rule{}, fmt.Errorf("ruleset: rule %q: %w", rule.ID, err)
	}
	rule.Params = params
	assert.That(rule.ID != "" && rule.Check.Kind != "", "a valid rule has an ID and a check kind")
	assert.That(len(rule.Languages) > 0, "a valid rule names one or more languages")
	assert.That(rule.Severity == "error" || rule.Severity == "warning" || rule.Severity == "note",
		"the schema allows three severities")
	return rule, nil
}

// decodeParams decodes the params struct of a rule into typed values.
func decodeParams(v cue.Value) (map[string]Param, error) {
	out := map[string]Param{}
	if !v.Exists() {
		return out, nil
	}
	assert.That(v.Err() == nil, "the validated params value has no error")
	iter, err := v.Fields()
	if err != nil {
		return nil, cueError("params", err)
	}
	for iter.Next() {
		if len(out) >= MaxParams {
			return nil, fmt.Errorf("more than %d parameters", MaxParams)
		}
		name := iter.Selector().Unquoted()
		p, err := decodeParam(iter.Value())
		if err != nil {
			return nil, fmt.Errorf("parameter %q: %w", name, err)
		}
		out[name] = p
	}
	assert.That(len(out) <= MaxParams, "the parameter count is within MaxParams")
	return out, nil
}

// decodeParam decodes one concrete parameter value. A default counts as concrete.
func decodeParam(v cue.Value) (Param, error) {
	assert.That(v.Exists(), "the parameter exists")
	if d, ok := v.Default(); ok {
		v = d
	}
	var p Param
	var err error
	switch v.Kind() {
	case cue.IntKind:
		p.Kind = ParamInt
		p.Int, err = v.Int64()
	case cue.StringKind:
		p.Kind = ParamString
		p.Str, err = v.String()
	case cue.BoolKind:
		p.Kind = ParamBool
		p.Bool, err = v.Bool()
	case cue.ListKind:
		p.Kind = ParamList
		err = v.Decode(&p.List)
	default:
		return Param{}, fmt.Errorf("unsupported kind %v", v.Kind())
	}
	if err != nil {
		return Param{}, cueError("decode parameter", err)
	}
	assert.That(p.Kind >= ParamInt && p.Kind <= ParamList, "the parameter has a valid kind")
	return p, nil
}
