package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// cli runs semscope with args and returns stdout, stderr and the exit code.
func cli(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if args == nil {
		args = []string{}
	}
	code := run(args, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

// golden compares got with testdata/<name>.golden.
// Set SEMSCOPE_UPDATE_GOLDEN=1 to write the file instead.
func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if os.Getenv("SEMSCOPE_UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path) //nolint:gosec // the test builds path from a fixed testdata name
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}
	if got != string(want) {
		t.Fatalf("output differs from %s:\n--- got\n%s--- want\n%s", path, got, want)
	}
}

func TestVersionFormsPrintTheSameText(t *testing.T) {
	long, _, code := cli(t, "--version")
	if code != exitOK || !strings.HasPrefix(long, "semscope ") || !strings.Contains(long, "sarif: 2.1.0, 2.2\n") {
		t.Fatalf("--version = %q, code %d", long, code)
	}
	for _, args := range [][]string{{"-v"}, {"version"}} {
		if out, _, code := cli(t, args...); out != long || code != exitOK {
			t.Errorf("%v = %q (code %d), want %q", args, out, code, long)
		}
	}
}

func TestUsageErrorsExitTwo(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"scan", "--format", "xml"}, `--format "xml"`},
		{[]string{"scan", "--sarif-version", "3.0"}, `--sarif-version "3.0"`},
		{[]string{"scan", "--omit-rule", "NOPE-01"}, `unknown rule "NOPE-01"`},
		{[]string{"scan", "--disable-rule", "NASA-P10-05", "--enable-rule", "NASA-P10-05"}, "in --enable-rule and in --omit-rule"},
		{[]string{"scan", "--set", "NASA-P10-04.max=1"}, `no parameter "max"`},
		{[]string{"scan", "-r", "missing.cue"}, "not a built-in ruleset"},
		{[]string{"scan", "--no-such-flag"}, "unknown flag"},
		{[]string{"rules", "inspect"}, "requires at least 1 arg"},
		{[]string{"version", "extra"}, "unknown command"},
	}
	for _, tc := range tests {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			_, stderr, code := cli(t, tc.args...)
			if code != exitUsage || !strings.Contains(stderr, tc.want) || strings.Count(stderr, "semscope:") != 1 {
				t.Fatalf("code %d, stderr %q, want exit 2 with one message that contains %q", code, stderr, tc.want)
			}
		})
	}
}

func TestScanReportsTheMissingEngine(t *testing.T) {
	_, stderr, code := cli(t, "scan", "--omit-rule", "NASA-P10-05", ".")
	if code != exitUsage || !strings.Contains(stderr, "9 active rule(s) for 1 path(s)") {
		t.Fatalf("code %d, stderr %q", code, stderr)
	}
}

func TestRulesListGolden(t *testing.T) {
	out, stderr, code := cli(t, "rules", "list")
	if code != exitOK {
		t.Fatalf("code %d, stderr %q", code, stderr)
	}
	golden(t, "rules-list", out)
}

func TestRulesInspectGolden(t *testing.T) {
	out, stderr, code := cli(t, "rules", "inspect", "nasa-p10", "../../testdata/custom_rulesets/add_custom.cue",
		"--omit-rule", "NASA-P10-05", "--set", "NASA-P10-04.max_lines=40")
	if code != exitOK {
		t.Fatalf("code %d, stderr %q", code, stderr)
	}
	golden(t, "rules-inspect", out)
}

func TestScanHelpGolden(t *testing.T) {
	out, _, code := cli(t, "scan", "--help")
	if code != exitOK {
		t.Fatalf("code %d", code)
	}
	golden(t, "scan-help", out)
}
