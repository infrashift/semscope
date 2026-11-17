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

// oracle runs the oracle of lang on the tree dir. goDirs are the directories
// of nasa-lint-go.py under dir; none means its default directories.
func oracle(t *testing.T, lang, dir string, goDirs ...string) outcome {
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
		cmd = exec.Command(python, append([]string{"-I", script, "--root", dir}, goDirs...)...) //nolint:gosec // fixed script path
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

// semscope scans the same files as the oracle of lang, with the parity
// overlay. paths are the scan paths; none means the default directories of
// the oracle.
func semscope(t *testing.T, lang, dir string, paths ...string) outcome {
	t.Helper()
	if len(paths) == 0 {
		for _, sub := range map[string][]string{"zig": {"src"}, "go": {"cmd", "internal", "rules"}}[lang] {
			if info, err := os.Stat(filepath.Join(dir, sub)); err == nil && info.IsDir() { //nolint:gosec // fixed fixture paths
				paths = append(paths, filepath.Join(dir, sub))
			}
		}
	}
	res, err := scan(t, load(t, []string{fixtures + "/parity.cue"}), DefaultLimits(), dir, paths...)
	if err != nil {
		t.Fatal(err)
	}
	o := outcome{findings: map[key]bool{}}
	for _, f := range res.Findings {
		if f.Path != "" && !strings.HasSuffix(f.Path, "."+lang) {
			continue // the oracle of lang reads only the files of lang
		}
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
func compare(t *testing.T, o, s outcome) {
	t.Helper()
	for _, d := range diff(o, s) {
		t.Errorf("only %s reports %v", d.Side, d.key)
	}
	if o.functions != s.functions || o.assertions != s.assertions || o.r5 != s.r5 {
		t.Errorf("R5: oracle %d functions, %d assertions, finding %v; semscope %d, %d, %v",
			o.functions, o.assertions, o.r5, s.functions, s.assertions, s.r5)
	}
}

// diff returns each finding that only one tool reports. Side is "oracle" or "semscope".
func diff(o, s outcome) []divergence {
	var out []divergence
	for k := range o.findings {
		if !s.findings[k] {
			out = append(out, divergence{key: k, Side: "oracle"})
		}
	}
	for k := range s.findings {
		if !o.findings[k] {
			out = append(out, divergence{key: k, Side: "semscope"})
		}
	}
	return out
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
			compare(t, o, s)
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
	compare(t, oracle(t, "go", root), semscope(t, "go", root))
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

// divergence is one recorded difference between an oracle and semscope on a
// corpus. A record has one of three forms:
//   - A line record has Path and Line: one finding that only Side reports.
//   - A class record has Count and no Path: the exact number of the other
//     findings of Rule that only Side reports. A class holds a known defect
//     with many instances, for example a false call edge of the oracle.
//   - A stats record has Side "stats": the R5 counts of both tools, as
//     [functions, assertions].
//
// Each record needs a reason and an issue. A record that the scan does not
// match exactly, also a stale record, fails the test.
type divergence struct {
	key
	Corpus   string  `json:"corpus"`
	Side     string  `json:"side"`
	Count    int     `json:"count,omitempty"`
	Oracle   *[2]int `json:"oracle,omitempty"`
	Semscope *[2]int `json:"semscope,omitempty"`
	Reason   string  `json:"reason"`
	Issue    string  `json:"issue"`
}

// corpus is one real source tree of the corpus parity test.
type corpus struct {
	lang string
	name string
	dir  string
}

// corpora reads the corpus lists: SEMSCOPE_PARITY_CORPUS holds Zig trees and
// SEMSCOPE_PARITY_GO_CORPUS holds Go trees. Each list uses the separator of
// the PATH variable. An entry is a directory or NAME=DIRECTORY. The default
// name is the base name of the directory. The records of
// parity-divergences.json use the name.
func corpora(t *testing.T) []corpus {
	t.Helper()
	var out []corpus
	for _, v := range []struct{ lang, env string }{{"zig", "SEMSCOPE_PARITY_CORPUS"}, {"go", "SEMSCOPE_PARITY_GO_CORPUS"}} {
		for _, entry := range filepath.SplitList(os.Getenv(v.env)) {
			name, dir, named := strings.Cut(entry, "=")
			if !named {
				dir = entry
			}
			if dir == "" {
				continue
			}
			abs, err := filepath.Abs(dir)
			if err != nil {
				t.Fatal(err)
			}
			if !named {
				name = filepath.Base(abs)
			}
			out = append(out, corpus{lang: v.lang, name: name, dir: abs})
		}
	}
	return out
}

// divergences reads parity-divergences.json and groups the records by corpus.
func divergences(t *testing.T) map[string][]divergence {
	t.Helper()
	data, err := os.ReadFile(fixtures + "/parity-divergences.json")
	if err != nil {
		t.Fatal(err)
	}
	var divs []divergence
	if err := json.Unmarshal(data, &divs); err != nil {
		t.Fatal(err)
	}
	out := map[string][]divergence{}
	for _, d := range divs {
		line, class, stats := d.Path != "" && d.Line > 0, d.Path == "" && d.Count > 0, d.Side == "stats"
		sideOK := (d.Side == "oracle" || d.Side == "semscope") != stats
		forms := 0
		for _, f := range []bool{line, class, stats && d.Oracle != nil && d.Semscope != nil} {
			if f {
				forms++
			}
		}
		if d.Corpus == "" || d.Reason == "" || d.Issue == "" || !sideOK || forms != 1 || (!stats && d.Rule == "") {
			t.Fatalf("divergence %+v needs a corpus, a reason, an issue and one form: line, class or stats", d)
		}
		out[d.Corpus] = append(out[d.Corpus], d)
	}
	return out
}

// checkRecords compares the differences of one corpus with its records.
func checkRecords(t *testing.T, o, s outcome, records []divergence) {
	t.Helper()
	lines := map[divergence]bool{}
	classes := map[[2]string]int{}
	var stats *divergence
	for i, d := range records {
		switch {
		case d.Side == "stats":
			stats = &records[i]
		case d.Count > 0:
			classes[[2]string{d.Rule, d.Side}] = d.Count
		default:
			lines[divergence{key: d.key, Side: d.Side}] = true
		}
	}
	got := map[[2]string]int{}
	for _, d := range diff(o, s) {
		if lines[d] {
			delete(lines, d)
			continue
		}
		got[[2]string{d.Rule, d.Side}]++
		if classes[[2]string{d.Rule, d.Side}] == 0 {
			t.Errorf("only %s reports %v", d.Side, d.key)
		}
	}
	for d := range lines {
		t.Errorf("stale line record: %s no longer reports %v alone", d.Side, d.key)
	}
	for c, n := range classes {
		if got[c] != n {
			t.Errorf("class record %s %s: %d differences, the record says %d", c[0], c[1], got[c], n)
		}
	}
	same := o.functions == s.functions && o.assertions == s.assertions && o.r5 == s.r5
	switch {
	case stats == nil && !same:
		t.Errorf("R5: oracle %d functions, %d assertions; semscope %d, %d; no stats record",
			o.functions, o.assertions, s.functions, s.assertions)
	case stats != nil && (*stats.Oracle != [2]int{o.functions, o.assertions} || *stats.Semscope != [2]int{s.functions, s.assertions} || o.r5 != s.r5):
		t.Errorf("R5: oracle %d functions, %d assertions; semscope %d, %d; the stats record says %v and %v",
			o.functions, o.assertions, s.functions, s.assertions, *stats.Oracle, *stats.Semscope)
	}
}

// zigTree returns a directory with a src/ tree for nasa-lint.py. A corpus
// without src/, for example the Zig standard library, is copied into src/
// of a temporary directory.
func zigTree(t *testing.T, dir string) string {
	t.Helper()
	if info, err := os.Stat(filepath.Join(dir, "src")); err == nil && info.IsDir() {
		return dir
	}
	tmp := t.TempDir()
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".zig") {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		dst := filepath.Join(tmp, "src", rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
			return err
		}
		data, err := os.ReadFile(p) //nolint:gosec // the corpus path comes from the test environment
		if err != nil {
			return err
		}
		return os.WriteFile(dst, data, 0o600) //nolint:gosec // dst is inside t.TempDir()
	})
	if err != nil {
		t.Fatal(err)
	}
	return tmp
}

// TestNasaParityCorpus compares the oracles and semscope on real source
// trees. SEMSCOPE_PARITY_CORPUS lists Zig trees, for example zmake, and
// SEMSCOPE_PARITY_GO_CORPUS lists Go trees, for example the cobra module.
// Each difference needs a record in parity-divergences.json.
func TestNasaParityCorpus(t *testing.T) {
	list := corpora(t)
	if len(list) == 0 {
		t.Skip("set SEMSCOPE_PARITY_CORPUS or SEMSCOPE_PARITY_GO_CORPUS to run the corpus parity test")
	}
	records := divergences(t)
	for _, c := range list {
		t.Run(c.lang+"/"+c.name, func(t *testing.T) {
			if c.lang == "zig" {
				root := zigTree(t, c.dir)
				checkRecords(t, oracle(t, "zig", root), semscope(t, "zig", root), records[c.name])
				return
			}
			checkRecords(t, oracle(t, "go", c.dir, "."), semscope(t, "go", c.dir, c.dir), records[c.name])
		})
	}
}
