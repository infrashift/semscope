package ruleset

import (
	"strings"
	"testing"
)

const custom = "../../testdata/custom_rulesets/"

func TestNasaP10Defaults(t *testing.T) {
	r := mustLoad(t, Options{Rulesets: []string{"nasa-p10"}})
	want := map[string]string{
		"NASA-P10-01": "no_goto_or_recursion", "NASA-P10-02": "bounded_loops",
		"NASA-P10-03": "no_dynamic_alloc", "NASA-P10-04": "max_function_lines",
		"NASA-P10-05": "min_assertions", "NASA-P10-06": "narrowest_scope",
		"NASA-P10-07": "checked_returns", "NASA-P10-08": "restricted_preprocessor",
		"NASA-P10-09": "max_pointer_indirection", "NASA-P10-10": "ast_query",
	}
	if len(r.Rules) != len(want) || len(r.Active()) != len(want) {
		t.Fatalf("rules = %s, active = %d", ids(r.Rules), len(r.Active()))
	}
	for _, rule := range r.Rules {
		if want[rule.ID] != rule.Check.Kind {
			t.Errorf("%s kind = %s, want %s", rule.ID, rule.Check.Kind, want[rule.ID])
		}
	}
	params := map[string]string{
		"NASA-P10-04.max_lines": "60", "NASA-P10-05.min_assertions": "2", "NASA-P10-05.mode": `"average"`,
		"NASA-P10-08.max_conditional_depth": "2", "NASA-P10-09.max_pointer_depth": "1",
		"NASA-P10-09.allow_function_pointers": "false",
	}
	for key, val := range params {
		id, name, _ := strings.Cut(key, ".")
		if got := findRule(t, r, id).Params[name].String(); got != val {
			t.Errorf("%s = %s, want %s", key, got, val)
		}
	}
	if findRule(t, r, "NASA-P10-03").Severity != "warning" || findRule(t, r, "NASA-P10-01").Severity != "error" {
		t.Errorf("severities: R3 is not warning or R1 is not error")
	}
	if q := findRule(t, r, "NASA-P10-10").Check.Query; len(q) != 6 || !strings.Contains(q["go"], "@violation") {
		t.Errorf("NASA-P10-10 queries = %v", q)
	}
}

func TestNasaP10AddCustomRule(t *testing.T) {
	r := mustLoad(t, Options{Rulesets: []string{"nasa-p10", custom + "add_custom.cue"}})
	if len(r.Active()) != 11 || findRule(t, r, "CUSTOM-01").Check.Kind != "token_ban" {
		t.Fatalf("rules = %s", ids(r.Active()))
	}
}

func TestNasaP10Omit05(t *testing.T) {
	for name, opts := range map[string]Options{
		"overlay": {Rulesets: []string{"nasa-p10", custom + "omit_p10_05.cue"}},
		"flag":    {Omit: []string{"NASA-P10-05"}},
	} {
		t.Run(name, func(t *testing.T) {
			r := mustLoad(t, opts)
			if got := ids(r.Omitted()); got != "NASA-P10-05" || len(r.Active()) != 9 {
				t.Fatalf("omitted = %s, active = %d", got, len(r.Active()))
			}
		})
	}
}

func TestNasaP10MaxLines40(t *testing.T) {
	for name, opts := range map[string]Options{
		"overlay": {Rulesets: []string{"nasa-p10", custom + "max_lines_40.cue"}},
		"flag":    {Set: []string{"NASA-P10-04.max_lines=40"}},
	} {
		t.Run(name, func(t *testing.T) {
			if got := findRule(t, mustLoad(t, opts), "NASA-P10-04").Params["max_lines"].Int; got != 40 {
				t.Fatalf("max_lines = %d, want 40", got)
			}
		})
	}
}

func TestNasaP10OverlayNarrowsLanguagesAndSeverity(t *testing.T) {
	r3 := findRule(t, mustLoad(t, Options{Rulesets: []string{"nasa-p10", custom + "alloc_c_only.cue"}}), "NASA-P10-03")
	if strings.Join(r3.Languages, ",") != "c" || r3.Severity != "error" {
		t.Fatalf("NASA-P10-03 = %v %s", r3.Languages, r3.Severity)
	}
}

func TestBuiltinNames(t *testing.T) {
	if got := strings.Join(BuiltinNames(), ","); got != "nasa-p10" {
		t.Fatalf("BuiltinNames() = %s", got)
	}
	if _, ok := Builtin("nope"); ok {
		t.Fatalf("Builtin(nope) = ok")
	}
}
