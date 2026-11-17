package report

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/infrashift/semscope/internal/engine"
	"github.com/infrashift/semscope/internal/parser"
	"github.com/infrashift/semscope/internal/ruleset"
)

func fixture() (engine.Result, []ruleset.Rule) {
	rules := []ruleset.Rule{
		{ID: "NASA-P10-02", Enabled: true, Severity: "error", Title: "Fixed loop bounds", Message: "Give each loop a fixed upper bound."},
		{ID: "NASA-P10-03", Enabled: true, Severity: "warning", Title: "No dynamic allocation", Message: "Do not allocate after init."},
		{ID: "NASA-P10-05", Enabled: true, Severity: "error", Title: "Assertion density", Message: "Use two assertions."},
		{ID: "NASA-P10-09", Enabled: false, Severity: "error", Title: "Omitted", Message: "Omitted."},
	}
	res := engine.Result{
		Files: 2,
		Stats: map[parser.Language]engine.LangStats{parser.Go: {Functions: 3, Assertions: 3}},
		Findings: []engine.Finding{
			{RuleID: "NASA-P10-05", Severity: "error", Message: "go: 3 assertions in 3 functions (average 1.00, min 2)"},
			{RuleID: "NASA-P10-02", Severity: "error", Message: "`for` loop without a condition", Path: "internal/a/a.go", Line: 8, Col: 2, EndLine: 10, EndCol: 3},
			{RuleID: "NASA-P10-03", Severity: "warning", Message: "allocation `make` in function `f`", Path: "b.go", Line: 4, Col: 9, EndLine: 4, EndCol: 22},
		},
	}
	return res, rules
}

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := "testdata/" + name + ".golden"
	if os.Getenv("SEMSCOPE_UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path) //nolint:gosec // fixed testdata path
	if err != nil || got != string(want) {
		t.Fatalf("output differs from %s (%v):\n--- got\n%s--- want\n%s", path, err, got, want)
	}
}

func TestSARIFGolden(t *testing.T) {
	res, rules := fixture()
	for _, v := range []string{"2.1.0", "2.2"} {
		t.Run(v, func(t *testing.T) {
			var b bytes.Buffer
			if err := WriteSARIF(&b, v, res, rules, "v1.0.0"); err != nil {
				t.Fatal(err)
			}
			var doc struct {
				Version string `json:"version"`
				Runs    []struct {
					Tool struct {
						Driver struct {
							Name  string            `json:"name"`
							Rules []json.RawMessage `json:"rules"`
						} `json:"driver"`
					} `json:"tool"`
					Results []json.RawMessage `json:"results"`
				} `json:"runs"`
			}
			if err := json.Unmarshal(b.Bytes(), &doc); err != nil {
				t.Fatal(err)
			}
			run := doc.Runs[0]
			if doc.Version != v || run.Tool.Driver.Name != "semscope" || len(run.Tool.Driver.Rules) != 3 || len(run.Results) != 3 {
				t.Fatalf("version %s, driver %s, %d rules, %d results", doc.Version, run.Tool.Driver.Name, len(run.Tool.Driver.Rules), len(run.Results))
			}
			golden(t, "report-"+v+".sarif", b.String())
		})
	}
}

func TestSARIFErrors(t *testing.T) {
	res, rules := fixture()
	var b bytes.Buffer
	if err := WriteSARIF(&b, "3.0", res, rules, "v1"); err == nil || !strings.Contains(err.Error(), "unsupported SARIF version") {
		t.Errorf("version 3.0: err = %v", err)
	}
	res.Findings = append(res.Findings, engine.Finding{RuleID: "NASA-P10-09", Severity: "error", Path: "c.go", Line: 1, Col: 1})
	if err := WriteSARIF(&b, "2.1.0", res, rules, "v1"); err == nil || !strings.Contains(err.Error(), "not active") {
		t.Errorf("omitted rule: err = %v", err)
	}
}

func TestTextAndJSONGolden(t *testing.T) {
	res, _ := fixture()
	var text, js bytes.Buffer
	if err := WriteText(&text, res); err != nil {
		t.Fatal(err)
	}
	if err := WriteJSON(&js, res, "v1.0.0"); err != nil {
		t.Fatal(err)
	}
	golden(t, "report.txt", text.String())
	golden(t, "report.json", js.String())
}
