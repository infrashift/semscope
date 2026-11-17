package ruleset

import (
	"fmt"
	"strings"
	"testing"
)

const td = "testdata/"

func mustLoad(t *testing.T, opts Options) Resolved {
	t.Helper()
	got, err := Load(opts)
	if err != nil {
		t.Fatalf("Load(%+v): %v", opts, err)
	}
	return got
}

func findRule(t *testing.T, r Resolved, id string) Rule {
	t.Helper()
	for _, rule := range r.Rules {
		if rule.ID == id {
			return rule
		}
	}
	t.Fatalf("rule %s not found in %v", id, ids(r.Rules))
	return Rule{}
}

func ids(rules []Rule) string {
	out := make([]string, len(rules))
	for i, r := range rules {
		out[i] = r.ID
	}
	return strings.Join(out, ",")
}

func TestLoadBaseResolvesDefaults(t *testing.T) {
	r := mustLoad(t, Options{Rulesets: []string{td + "base.cue"}})
	if got := ids(r.Rules); got != "TEST-01,TEST-02,TEST-03" {
		t.Fatalf("rules = %s", got)
	}
	r1 := findRule(t, r, "TEST-01")
	if !r1.Enabled || r1.Severity != "error" || r1.Params["max_lines"].Int != 60 {
		t.Fatalf("TEST-01 = %+v", r1)
	}
	r2 := findRule(t, r, "TEST-02")
	if r2.Params["mode"].Str != "average" || r2.Params["strict"].Kind != ParamBool {
		t.Fatalf("TEST-02 params = %+v", r2.Params)
	}
	if got := r2.Check.AssertCalls["go"]; len(got) != 1 || got[0] != "assert.That" {
		t.Fatalf("TEST-02 assert_calls = %v", got)
	}
	if findRule(t, r, "TEST-03").Severity != "warning" {
		t.Fatalf("TEST-03 severity is not warning")
	}
	if len(r.Sources) != 1 || r.Sources[0] != td+"base.cue" {
		t.Fatalf("sources = %v", r.Sources)
	}
}

func TestOverlayAddsCustomRule(t *testing.T) {
	r := mustLoad(t, Options{Rulesets: []string{td + "base.cue", td + "custom.cue"}})
	c := findRule(t, r, "CUSTOM-01")
	if c.Check.Kind != "token_ban" || c.Check.Patterns["go"][0] != `fmt\.Println` || !c.Enabled {
		t.Fatalf("CUSTOM-01 = %+v", c)
	}
}

func TestOmitRule(t *testing.T) {
	for name, opts := range map[string]Options{
		"overlay": {Rulesets: []string{td + "base.cue", td + "omit.cue"}},
		"flag":    {Rulesets: []string{td + "base.cue"}, Omit: []string{"TEST-02"}},
	} {
		t.Run(name, func(t *testing.T) {
			r := mustLoad(t, opts)
			if got := ids(r.Active()); got != "TEST-01,TEST-03" {
				t.Fatalf("active = %s", got)
			}
			if got := ids(r.Omitted()); got != "TEST-02" {
				t.Fatalf("omitted = %s", got)
			}
		})
	}
}

func TestEnableRuleFlagOverridesOverlay(t *testing.T) {
	r := mustLoad(t, Options{Rulesets: []string{td + "base.cue", td + "omit.cue"}, Enable: []string{"TEST-02"}})
	if !findRule(t, r, "TEST-02").Enabled {
		t.Fatalf("TEST-02 is not enabled")
	}
}

func TestSetMaxLines(t *testing.T) {
	for name, opts := range map[string]Options{
		"overlay": {Rulesets: []string{td + "base.cue", td + "maxlines.cue"}},
		"flag":    {Rulesets: []string{td + "base.cue"}, Set: []string{"TEST-01.max_lines=40"}},
	} {
		t.Run(name, func(t *testing.T) {
			if got := findRule(t, mustLoad(t, opts), "TEST-01").Params["max_lines"].Int; got != 40 {
				t.Fatalf("max_lines = %d, want 40", got)
			}
		})
	}
}

func TestSetTypedValues(t *testing.T) {
	r := mustLoad(t, Options{Rulesets: []string{td + "base.cue"}, Set: []string{
		"TEST-02.strict=true", "TEST-02.mode=per_function", "TEST-02.allow= init, setup ,",
	}})
	p := findRule(t, r, "TEST-02").Params
	if !p["strict"].Bool || p["mode"].Str != "per_function" || strings.Join(p["allow"].List, "|") != "init|setup" {
		t.Fatalf("params = %+v", p)
	}
}

func TestLoadErrors(t *testing.T) {
	base := td + "base.cue"
	tests := []struct {
		name string
		opts Options
		want string
	}{
		{"unknown source", Options{Rulesets: []string{td + "missing.cue"}}, "not a built-in ruleset"},
		{"typo field", Options{Rulesets: []string{base, td + "typo.cue"}}, "field not allowed"},
		{"bad kind", Options{Rulesets: []string{base, td + "badkind.cue"}}, "BAD-01"},
		{"missing param", Options{Rulesets: []string{base, td + "noparam.cue"}}, "max_lines"},
		{"bad language", Options{Rulesets: []string{base, td + "badlang.cue"}}, "BAD-03"},
		{"concrete conflict", Options{Rulesets: []string{base, td + "conflict.cue"}}, "conflicting values"},
		{"set unknown rule", Options{Rulesets: []string{base}, Set: []string{"NOPE-01.x=1"}}, `unknown rule "NOPE-01"`},
		{"set unknown param", Options{Rulesets: []string{base}, Set: []string{"TEST-01.max=1"}}, `no parameter "max"`},
		{"set bad int", Options{Rulesets: []string{base}, Set: []string{"TEST-01.max_lines=ten"}}, "want an integer"},
		{"set bad bool", Options{Rulesets: []string{base}, Set: []string{"TEST-02.strict=maybe"}}, "want true or false"},
		{"set bad enum", Options{Rulesets: []string{base}, Set: []string{"TEST-02.mode=sum"}}, "TEST-02"},
		{"set no equals", Options{Rulesets: []string{base}, Set: []string{"TEST-01.max_lines"}}, "want RULE_ID.param=value"},
		{"set no dot", Options{Rulesets: []string{base}, Set: []string{"TEST-01=4"}}, "want RULE_ID.param=value"},
		{"set concrete", Options{Rulesets: []string{base}, Set: []string{"TEST-03.label=x"}}, "conflicting values"},
		{"enable unknown", Options{Rulesets: []string{base}, Enable: []string{"NOPE-01"}}, "--enable-rule"},
		{"omit unknown", Options{Rulesets: []string{base}, Omit: []string{"NOPE-01"}}, "--omit-rule"},
		{"enable and omit", Options{Rulesets: []string{base}, Enable: []string{"TEST-01"}, Omit: []string{"TEST-01"}}, "in --enable-rule and in --omit-rule"},
		{"too many sets", Options{Rulesets: []string{base}, Set: make([]string, MaxOverrides+1)}, "--set values given"},
		{"not CUE", Options{Rulesets: []string{td + "dir/notes.txt"}}, "notes.txt"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Load(tc.opts)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Load() error = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

func TestEmptyDirectoryGivesError(t *testing.T) {
	dir := t.TempDir()
	if _, err := Load(Options{Rulesets: []string{dir}}); err == nil || !strings.Contains(err.Error(), "no .cue files") {
		t.Fatalf("Load(empty dir) error = %v", err)
	}
}

func TestDirectorySourceReadsCueFilesInNameSequence(t *testing.T) {
	r := mustLoad(t, Options{Rulesets: []string{td + "dir"}})
	if got := strings.Join(r.Sources, ","); got != td+"dir/a.cue,"+td+"dir/b.cue" {
		t.Fatalf("sources = %s", got)
	}
	if findRule(t, r, "DIR-01").Severity != "note" {
		t.Fatalf("b.cue did not set the severity")
	}
}

func TestLoadWithoutRulesetsReadsDefault(t *testing.T) {
	r := mustLoad(t, Options{})
	if len(r.Sources) != 1 || r.Sources[0] != "builtin:"+DefaultRuleset {
		t.Fatalf("sources = %v", r.Sources)
	}
}

func TestParamString(t *testing.T) {
	for _, tc := range []struct {
		p    Param
		want string
	}{
		{Param{Kind: ParamInt, Int: 60}, "60"},
		{Param{Kind: ParamString, Str: "average"}, `"average"`},
		{Param{Kind: ParamBool, Bool: true}, "true"},
		{Param{Kind: ParamList, List: []string{"a", "b"}}, `["a", "b"]`},
	} {
		if got := tc.p.String(); got != tc.want {
			t.Errorf("%+v.String() = %s, want %s", tc.p, got, tc.want)
		}
	}
}

func ExampleLoad() {
	r, err := Load(Options{Rulesets: []string{"testdata/base.cue"}, Set: []string{"TEST-01.max_lines=40"}})
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, rule := range r.Active() {
		fmt.Println(rule.ID, rule.Check.Kind)
	}
	// Output:
	// TEST-01 max_function_lines
	// TEST-02 min_assertions
	// TEST-03 token_ban
}
