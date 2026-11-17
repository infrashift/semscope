package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/infrashift/semscope/internal/parser"
)

// The parity tests of ADR-0005: scripts/nasa-lint.py (Zig) and
// scripts/nasa-lint-go.py (Go) against semscope with testdata/nasa_p10/parity.cue.

const fixtures = "../../testdata/nasa_p10"

var (
	oracleLine    = regexp.MustCompile(`^(\S+\.(?:zig|go)):(\d+): R(\d+):`)
	oracleSummary = regexp.MustCompile(`(\d+) functions, (\d+) assertions`)
)

// key identifies one finding for the comparison.
type key struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	Rule string `json:"rule"`
}

// outcome is what one tool reports for one tree.
type outcome struct {
	findings   map[key]bool
	functions  int
	assertions int
	r5         bool
}

// oracle runs the oracle of lang on the tree dir.
func oracle(t *testing.T, lang, dir string) outcome {
	t.Helper()
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Fatalf("the parity tests need python3 (ADR-0005): %v", err)
	}
	script, err := filepath.Abs("../../scripts/nasa-lint.py")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(python, "-I", script) //nolint:gosec // fixed script path
	cmd.Dir = dir
	if lang == "go" {
		script = strings.Replace(script, "nasa-lint.py", "nasa-lint-go.py", 1)
		cmd = exec.Command(python, "-I", script, "--root", dir) //nolint:gosec // fixed script path
	}
	out, err := cmd.Output()
	var exit *exec.ExitError
	if err != nil && (!errors.As(err, &exit) || exit.ExitCode() != 1) {
		t.Fatalf("%s oracle on %s: %v", lang, dir, err)
	}
	return parseOracle(t, string(out))
}

// parseOracle reads the finding lines and the summary line of an oracle.
func parseOracle(t *testing.T, out string) outcome {
	t.Helper()
	o := outcome{findings: map[key]bool{}}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, l := range lines {
		if m := oracleLine.FindStringSubmatch(l); m != nil {
			o.findings[key{m[1], atoi(t, m[2]), fmt.Sprintf("NASA-P10-%02d", atoi(t, m[3]))}] = true
		}
		o.r5 = o.r5 || strings.Contains(l, ": R5: ")
	}
	m := oracleSummary.FindStringSubmatch(lines[len(lines)-1])
	if m == nil {
		t.Fatalf("no oracle summary line in:\n%s", out)
	}
	o.functions, o.assertions = atoi(t, m[1]), atoi(t, m[2])
	return o
}

// atoi parses a decimal number from an oracle line.
func atoi(t *testing.T, s string) int {
	t.Helper()
	n, err := strconv.Atoi(s)
	if err != nil {
		t.Fatalf("oracle number %q: %v", s, err)
	}
	return n
}

// semscope scans the same files as the oracle of lang, with the parity overlay.
func semscope(t *testing.T, lang, dir string) outcome {
	t.Helper()
	var paths []string
	for _, sub := range map[string][]string{"zig": {"src"}, "go": {"cmd", "internal", "rules"}}[lang] {
		if info, err := os.Stat(filepath.Join(dir, sub)); err == nil && info.IsDir() { //nolint:gosec // fixed fixture paths
			paths = append(paths, filepath.Join(dir, sub))
		}
	}
	res, err := scan(t, load(t, []string{fixtures + "/parity.cue"}), DefaultLimits(), dir, paths...)
	if err != nil {
		t.Fatal(err)
	}
	o := outcome{findings: map[key]bool{}}
	for _, f := range res.Findings {
		if f.Path == "" {
			o.r5 = o.r5 || f.RuleID == "NASA-P10-05"
			continue
		}
		o.findings[key{f.Path, f.Line, f.RuleID}] = true
	}
	st := res.Stats[parser.Language(lang)]
	o.functions, o.assertions = st.Functions, st.Assertions
	return o
}

// compare reports each difference between the oracle and semscope, except
// the differences in allowed.
func compare(t *testing.T, o, s outcome, allowed map[key]bool) {
	t.Helper()
	for k := range o.findings {
		if !s.findings[k] && !allowed[k] {
			t.Errorf("only the oracle reports %v", k)
		}
	}
	for k := range s.findings {
		if !o.findings[k] && !allowed[k] {
			t.Errorf("only semscope reports %v", k)
		}
	}
	if o.functions != s.functions || o.assertions != s.assertions || o.r5 != s.r5 {
		t.Errorf("R5: oracle %d functions, %d assertions, finding %v; semscope %d, %d, %v",
			o.functions, o.assertions, o.r5, s.functions, s.assertions, s.r5)
	}
}

func TestNasaParityFixtures(t *testing.T) {
	dirs, err := filepath.Glob(fixtures + "/violations/R*/*")
	if err != nil {
		t.Fatal(err)
	}
	dirs = append(dirs, fixtures+"/compliant/zig", fixtures+"/compliant/go")
	cases := 0
	oracleRules := map[string]bool{"R01": true, "R02": true, "R04": true, "R05": true, "R07": true, "R09": true, "compliant": true}
	for _, dir := range dirs {
		lang := filepath.Base(dir)
		if (lang != "zig" && lang != "go") || !oracleRules[filepath.Base(filepath.Dir(dir))] {
			continue
		}
		cases++
		t.Run(strings.TrimPrefix(dir, fixtures+"/"), func(t *testing.T) {
			abs, err := filepath.Abs(dir)
			if err != nil {
				t.Fatal(err)
			}
			o, s := oracle(t, lang, abs), semscope(t, lang, abs)
			compare(t, o, s, nil)
			checkFixtureFires(t, dir, o)
		})
	}
	if cases != 14 {
		t.Fatalf("%d parity cases, want 14 (6 rules and the compliant tree, for Zig and Go)", cases)
	}
}

// checkFixtureFires checks that a violation fixture triggers its own rule and
// that a compliant fixture triggers nothing.
func checkFixtureFires(t *testing.T, dir string, o outcome) {
	t.Helper()
	if strings.Contains(dir, "/compliant/") {
		if len(o.findings) != 0 || o.r5 {
			t.Errorf("the compliant fixture has findings: %v", o.findings)
		}
		return
	}
	rule := "NASA-P10-" + strings.TrimPrefix(filepath.Base(filepath.Dir(dir)), "R")
	for k := range o.findings {
		if k.Rule == rule {
			return
		}
	}
	if rule != "NASA-P10-05" || !o.r5 {
		t.Errorf("the fixture does not trigger %s", rule)
	}
}

func TestNasaParitySemscopeSources(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	compare(t, oracle(t, "go", root), semscope(t, "go", root), nil)
}

func TestSingleRuleExpected(t *testing.T) {
	for n := 1; n <= 10; n++ {
		id := fmt.Sprintf("NASA-P10-%02d", n)
		dir := fmt.Sprintf("%s/violations/R%02d", fixtures, n)
		t.Run(id, func(t *testing.T) {
			rules := load(t, nil)
			for i := range rules {
				rules[i].Enabled = rules[i].ID == id
			}
			res, err := scan(t, rules, DefaultLimits(), dir, dir)
			if err != nil {
				t.Fatal(err)
			}
			checkExpected(t, dir+"/expected.txt", text(t, res))
		})
	}
}

// checkExpected compares the finding lines of got with the file path. The
// summary lines of the two reports differ in form, so the check skips them.
func checkExpected(t *testing.T, path, got string) {
	t.Helper()
	want, err := os.ReadFile(path) //nolint:gosec // fixed testdata path
	if err != nil {
		t.Fatal(err)
	}
	wantLines := strings.Split(strings.TrimSpace(string(want)), "\n")
	gotLines := strings.Split(strings.TrimSpace(got), "\n")
	if strings.Join(wantLines[:len(wantLines)-1], "\n") != strings.Join(gotLines[:len(gotLines)-1], "\n") {
		t.Fatalf("findings differ from %s:\n--- got\n%s--- want\n%s", path, got, want)
	}
}

// divergence is one recorded difference between an oracle and semscope on a corpus.
type divergence struct {
	key
	Side   string `json:"side"`
	Reason string `json:"reason"`
	Issue  string `json:"issue"`
}

// TestNasaParityCorpus compares nasa-lint.py and semscope on a real Zig tree.
// Set SEMSCOPE_PARITY_CORPUS to a directory with a src/ tree, for example zmake.
func TestNasaParityCorpus(t *testing.T) {
	corpus := os.Getenv("SEMSCOPE_PARITY_CORPUS")
	if corpus == "" {
		t.Skip("set SEMSCOPE_PARITY_CORPUS to run the corpus parity test")
	}
	data, err := os.ReadFile(fixtures + "/parity-divergences.json")
	if err != nil {
		t.Fatal(err)
	}
	var divs []divergence
	if err := json.Unmarshal(data, &divs); err != nil {
		t.Fatal(err)
	}
	allowed := map[key]bool{}
	for _, d := range divs {
		if d.Reason == "" || d.Issue == "" {
			t.Fatalf("divergence %v needs a reason and an issue", d.key)
		}
		allowed[d.key] = true
	}
	abs, err := filepath.Abs(corpus)
	if err != nil {
		t.Fatal(err)
	}
	compare(t, oracle(t, "zig", abs), semscope(t, "zig", abs), allowed)
}
