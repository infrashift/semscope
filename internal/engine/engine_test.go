package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/infrashift/semscope/internal/ruleset"
)

// load returns the enabled rules of nasa-p10 plus the extra overlays and --set values.
func load(t *testing.T, overlays []string, set ...string) []ruleset.Rule {
	t.Helper()
	rs, err := ruleset.Load(ruleset.Options{Rulesets: append([]string{"nasa-p10"}, overlays...), Set: set})
	if err != nil {
		t.Fatal(err)
	}
	return rs.Rules
}

// scan runs one scan of paths with rules and limits.
func scan(t *testing.T, rules []ruleset.Rule, limits Limits, root string, paths ...string) (Result, error) {
	t.Helper()
	e, err := New(rules, limits)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	return e.Scan(root, paths)
}

// text renders res in the same form as the text report.
func text(t *testing.T, res Result) string {
	t.Helper()
	var b strings.Builder
	for _, f := range res.Findings {
		where := "-"
		if f.Path != "" {
			where = fmt.Sprintf("%s:%d:%d", f.Path, f.Line, f.Col)
		}
		fmt.Fprintf(&b, "%s: %s: %s: %s\n", where, f.Severity, f.RuleID, f.Message)
	}
	fmt.Fprintf(&b, "%d finding(s) in %d file(s)\n", len(res.Findings), res.Files)
	return b.String()
}

// tree writes files under a new temporary directory and returns it.
func tree(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestSamplesGolden(t *testing.T) {
	samples, err := filepath.Glob("../parser/testdata/sample_*")
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, s := range samples {
		if !strings.HasSuffix(s, ".golden") {
			paths = append(paths, s)
		}
	}
	res, err := scan(t, load(t, nil), DefaultLimits(), "../parser/testdata", paths...)
	if err != nil {
		t.Fatal(err)
	}
	got := text(t, res)
	golden := "testdata/samples.golden"
	if os.Getenv("SEMSCOPE_UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(golden, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil || got != string(want) {
		t.Fatalf("scan output differs from %s (%v):\n--- got\n%s--- want\n%s", golden, err, got, want)
	}
	again, err := scan(t, load(t, nil), DefaultLimits(), "../parser/testdata", paths...)
	if err != nil || text(t, again) != got {
		t.Fatalf("a second scan gives other output (%v)", err)
	}
}

func TestSuppressionTags(t *testing.T) {
	dir := tree(t, map[string]string{"a.go": `package a

func f(ch chan int) {
	// semscope: allow(NASA-P10-02) the loop ends when ch closes
	for {
		<-ch
	}
}

func g(ch chan int) {
	for { // nasa-lint: allow(R2) same reason
		<-ch
	}
}

func h(ch chan int) {
	for {
		<-ch
	}
}
`})
	res, err := scan(t, load(t, nil), DefaultLimits(), dir, dir)
	if err != nil {
		t.Fatal(err)
	}
	var loops []string
	for _, f := range res.Findings {
		if f.RuleID == "NASA-P10-02" {
			loops = append(loops, f.Path+":"+strconv.Itoa(f.Line))
		}
	}
	if strings.Join(loops, ",") != "a.go:17" {
		t.Fatalf("NASA-P10-02 findings = %v, want only a.go:17", loops)
	}
}

func TestCustomRules(t *testing.T) {
	overlay := filepath.Join(t.TempDir(), "q.cue")
	query := `package overlay
rules: "CUSTOM-02": {
	languages: ["python"]
	title: "No print"
	message: "Use logging."
	check: {kind: "ast_query", query: python: "(call function: (identifier) @violation (#eq? @violation \"print\"))"}
}
`
	if err := os.WriteFile(overlay, []byte(query), 0o600); err != nil {
		t.Fatal(err)
	}
	dir := tree(t, map[string]string{
		"a.go": "package a\n\nimport \"fmt\"\n\nfunc f() {\n\tfmt.Println(\"fmt.Println(\")\n}\n",
		"b.py": "def f():\n    print('x')\n",
	})
	rules := load(t, []string{"../../testdata/custom_rulesets/add_custom.cue", overlay})
	res, err := scan(t, rules, DefaultLimits(), dir, dir)
	if err != nil {
		t.Fatal(err)
	}
	out := text(t, res)
	if !strings.Contains(out, "a.go:6:2: error: CUSTOM-01:") || strings.Count(out, "CUSTOM-01") != 1 {
		t.Fatalf("CUSTOM-01 (token_ban, outside strings only):\n%s", out)
	}
	if !strings.Contains(out, "b.py:2:5: error: CUSTOM-02: construct `print`") {
		t.Fatalf("CUSTOM-02 (ast_query):\n%s", out)
	}
}

func TestPerFunctionAssertions(t *testing.T) {
	dir := tree(t, map[string]string{"a.py": "def f(x):\n    assert x\n    assert x\n    assert x\n    assert x\n\ndef g(x):\n    return x\n"})
	avg, err := scan(t, load(t, nil), DefaultLimits(), dir, dir)
	if err != nil || strings.Contains(text(t, avg), "NASA-P10-05") {
		t.Fatalf("average mode: 4 assertions in 2 functions passes (%v):\n%s", err, text(t, avg))
	}
	per, err := scan(t, load(t, nil, "NASA-P10-05.mode=per_function"), DefaultLimits(), dir, dir)
	if err != nil || !strings.Contains(text(t, per), "a.py:7:1: error: NASA-P10-05: function `g` has 0 assertion(s) (min 2)") {
		t.Fatalf("per_function mode (%v):\n%s", err, text(t, per))
	}
}

func TestLimitsAndCompileErrors(t *testing.T) {
	dir := tree(t, map[string]string{"a.go": "package a\nfunc f() {\n\tfor {\n\t}\n}\n", "b.go": "package a\n"})
	few := DefaultLimits()
	few.MaxFiles = 1
	if _, err := scan(t, load(t, nil), few, dir, dir); err == nil || !strings.Contains(err.Error(), "more than 1 files") {
		t.Errorf("MaxFiles: err = %v", err)
	}
	small := DefaultLimits()
	small.MaxFindings = 1
	if _, err := scan(t, load(t, nil), small, dir, dir); err == nil || !strings.Contains(err.Error(), "max 1") {
		t.Errorf("MaxFindings: err = %v", err)
	}
	if _, err := New(load(t, nil), Limits{}); err == nil {
		t.Errorf("New(Limits{}) gave no error")
	}
	bad := load(t, nil)
	bad[9].Check.Query = map[string]string{"go": "(nope) @violation"}
	if _, err := New(bad, DefaultLimits()); err == nil || !strings.Contains(err.Error(), "NASA-P10-10") {
		t.Errorf("bad query: err = %v", err)
	}
	bad[9].Check.Query = map[string]string{"go": "(comment) @other"}
	if _, err := New(bad, DefaultLimits()); err == nil || !strings.Contains(err.Error(), "@violation") {
		t.Errorf("query without @violation: err = %v", err)
	}
}

func TestExclude(t *testing.T) {
	dir := tree(t, map[string]string{
		"keep/a.go":       "package a\nfunc f() {\n\tfor {\n\t}\n}\n",
		"tools/b.go":      "package b\nfunc f() {\n\tfor {\n\t}\n}\n",
		"gen/x/c_gen.go":  "package c\nfunc f() {\n\tfor {\n\t}\n}\n",
		"gen/x/c_real.go": "package c\nfunc g() {\n\tfor {\n\t}\n}\n",
	})
	e, err := New(load(t, nil), DefaultLimits(), "tools", "gen/*/*_gen.go")
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	res, err := e.Scan(dir, []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if res.Files != 2 || !strings.Contains(text(t, res), "keep/a.go") || !strings.Contains(text(t, res), "gen/x/c_real.go") {
		t.Fatalf("files = %d:\n%s", res.Files, text(t, res))
	}
	if _, err := New(load(t, nil), DefaultLimits(), "["); err == nil {
		t.Fatalf("a bad exclude pattern gave no error")
	}
}

func TestNasaTag(t *testing.T) {
	for id, want := range map[string]string{"NASA-P10-01": "R1", "NASA-P10-10": "R10", "NASA-P10-11": "", "CUSTOM-01": ""} {
		if got := nasaTag(id); got != want {
			t.Errorf("nasaTag(%s) = %q, want %q", id, got, want)
		}
	}
}

func TestGoTypesCheck(t *testing.T) {
	dir := "../../testdata/go_types"
	rules := load(t, nil, "NASA-P10-07.go_types=true")
	for i := range rules {
		rules[i].Enabled = rules[i].ID == "NASA-P10-07"
	}
	res, err := scan(t, rules, DefaultLimits(), dir, dir)
	if err != nil {
		t.Fatal(err)
	}
	want := `drops.go:32:2: error: NASA-P10-07: unchecked error of ` + "`example.com/gotypes.fail`" + `
drops.go:33:2: error: NASA-P10-07: unchecked error of ` + "`example.com/gotypes.pair`" + `
drops.go:34:8: error: NASA-P10-07: unchecked error of ` + "`(example.com/gotypes.closer).Close` in a defer statement" + `
drops.go:35:5: error: NASA-P10-07: unchecked error of ` + "`example.com/gotypes.fail` in a go statement" + `
drops.go:36:2: error: NASA-P10-07: unchecked error of ` + "`example.com/gotypes.codedFail`" + `
drops.go:42:2: error: NASA-P10-07: unchecked error of ` + "`fmt.Fprintln`" + `
drops.go:45:2: error: NASA-P10-07: unchecked error of ` + "`f`" + `
7 finding(s) in 2 file(s)
`
	if got := text(t, res); got != want {
		t.Fatalf("findings:\n%s\nwant:\n%s", got, want)
	}
	off, err := scan(t, load(t, nil), DefaultLimits(), dir, dir+"/drops.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range off.Findings {
		if strings.HasPrefix(f.Message, "unchecked error") {
			t.Errorf("go_types is false by default, but the scan reports %v", f)
		}
	}
}
