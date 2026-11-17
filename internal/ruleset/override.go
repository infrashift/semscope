package ruleset

import (
	"fmt"
	"strconv"
	"strings"

	"cuelang.org/go/cue"

	"github.com/infrashift/semscope/internal/assert"
)

// applySets fills each RULE_ID.param=value override into v.
// The parameter must exist. The value takes the type of the parameter.
func applySets(v cue.Value, sets []string) (cue.Value, error) {
	assert.That(v.Exists(), "v exists")
	assert.That(len(sets) <= MaxOverrides, "the override count is within MaxOverrides")
	for _, s := range sets {
		id, param, raw, err := parseSet(s)
		if err != nil {
			return cue.Value{}, err
		}
		if !v.LookupPath(cue.MakePath(cue.Str("rules"), cue.Str(id))).Exists() {
			return cue.Value{}, fmt.Errorf("ruleset: --set %s: unknown rule %q", s, id)
		}
		path := cue.MakePath(cue.Str("rules"), cue.Str(id), cue.Str("params"), cue.Str(param))
		cur := v.LookupPath(path)
		if !cur.Exists() {
			return cue.Value{}, fmt.Errorf("ruleset: --set %s: rule %q has no parameter %q", s, id, param)
		}
		x, err := parseParamValue(cur, raw)
		if err != nil {
			return cue.Value{}, fmt.Errorf("ruleset: --set %s: %w", s, err)
		}
		v = v.FillPath(path, x)
		if err := v.LookupPath(path).Err(); err != nil {
			return cue.Value{}, cueError("--set "+s, err)
		}
	}
	return v, nil
}

// parseSet splits RULE_ID.param=value into its three parts.
func parseSet(s string) (id, param, value string, err error) {
	lhs, value, ok := strings.Cut(s, "=")
	if !ok {
		return "", "", "", fmt.Errorf("ruleset: --set %q: want RULE_ID.param=value", s)
	}
	id, param, ok = strings.Cut(strings.TrimSpace(lhs), ".")
	if !ok || id == "" || param == "" {
		return "", "", "", fmt.Errorf("ruleset: --set %q: want RULE_ID.param=value", s)
	}
	assert.That(id != "" && param != "", "a parsed override has an ID and a parameter")
	assert.That(!strings.Contains(id, "."), "the ID ends at the first dot")
	return id, param, strings.TrimSpace(value), nil
}

// parseParamValue parses raw with the type of the current parameter value cur.
// A list value is comma-separated. An empty raw value gives an empty list.
func parseParamValue(cur cue.Value, raw string) (any, error) {
	assert.That(cur.Exists(), "the parameter exists")
	assert.That(raw == strings.TrimSpace(raw), "parseSet trims the value")
	kind := cur.IncompleteKind()
	switch kind {
	case cue.IntKind:
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("want an integer, got %q", raw)
		}
		return n, nil
	case cue.BoolKind:
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("want true or false, got %q", raw)
		}
		return b, nil
	case cue.StringKind:
		return raw, nil
	case cue.ListKind:
		items := []string{}
		for _, item := range strings.Split(raw, ",") {
			if item = strings.TrimSpace(item); item != "" {
				items = append(items, item)
			}
		}
		return items, nil
	}
	return nil, fmt.Errorf("the parameter has no single type (%v)", kind)
}

// applyToggles applies the --enable-rule and --omit-rule lists to rules.
// An unknown ID, or an ID in both lists, gives an error.
func applyToggles(rules []Rule, enable, omit []string) error {
	assert.That(len(rules) <= MaxRules, "the rule count is within MaxRules")
	index := make(map[string]int, len(rules))
	for i, r := range rules {
		index[r.ID] = i
	}
	enabled := make(map[string]bool, len(enable))
	for _, id := range enable {
		i, ok := index[id]
		if !ok {
			return fmt.Errorf("ruleset: --enable-rule: unknown rule %q", id)
		}
		rules[i].Enabled = true
		enabled[id] = true
	}
	for _, id := range omit {
		i, ok := index[id]
		if !ok {
			return fmt.Errorf("ruleset: --omit-rule: unknown rule %q", id)
		}
		if enabled[id] {
			return fmt.Errorf("ruleset: rule %q is in --enable-rule and in --omit-rule", id)
		}
		rules[i].Enabled = false
	}
	assert.That(len(index) == len(rules), "rule IDs are unique")
	return nil
}
