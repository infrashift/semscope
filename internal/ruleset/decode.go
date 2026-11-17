package ruleset

import (
	"fmt"
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

// decodeRule decodes one concrete rule.
func decodeRule(v cue.Value) (Rule, error) {
	var rule Rule
	if err := v.Decode(&rule); err != nil {
		return Rule{}, cueError("decode rule", err)
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
