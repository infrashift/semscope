package main

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/infrashift/semscope/internal/assert"
	"github.com/infrashift/semscope/internal/engine"
	"github.com/infrashift/semscope/internal/report"
	"github.com/infrashift/semscope/internal/ruleset"
	"github.com/infrashift/semscope/internal/version"
)

// scanFlags holds the flags of `semscope scan`.
type scanFlags struct {
	rules        rulesetFlags
	format       string
	sarifVersion string
	output       string
	failLevel    string
}

func newScanCmd() *cobra.Command {
	f := &scanFlags{}
	cmd := &cobra.Command{
		Use:   "scan [paths...]",
		Short: "Scan source files with the active rules",
		Long: `Scan the source files under each path with the active rules of the ruleset.
The default path is the current directory.
The exit status is 1 when a finding has the --fail-level severity or a higher severity.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScan(cmd, f, args)
		},
	}
	fs := cmd.Flags()
	f.rules.bind(fs, true)
	fs.StringVar(&f.format, "format", "text", "the output format: text, json or sarif")
	fs.StringVar(&f.sarifVersion, "sarif-version", version.DefaultSARIFVersion,
		"the SARIF version: 2.1.0, or 2.2 (an OASIS draft)")
	fs.StringVarP(&f.output, "output", "o", "", "the output file (default stdout)")
	fs.StringVar(&f.failLevel, "fail-level", "error",
		"the lowest severity that gives exit status 1: error, warning, note or none")
	assert.That(fs.Lookup("format") != nil, "the --format flag exists")
	assert.That(cmd.RunE != nil, "the scan command runs")
	return cmd
}

// validate checks the values of the output flags.
func (f *scanFlags) validate() error {
	switch f.format {
	case "text", "json", "sarif":
	default:
		return fmt.Errorf("--format %q: want text, json or sarif", f.format)
	}
	switch f.failLevel {
	case "error", "warning", "note", "none":
	default:
		return fmt.Errorf("--fail-level %q: want error, warning, note or none", f.failLevel)
	}
	if !version.IsSupportedSARIF(f.sarifVersion) {
		return fmt.Errorf("--sarif-version %q: want 2.1.0 or 2.2", f.sarifVersion)
	}
	assert.That(f.format != "", "the format has a value")
	assert.That(f.sarifVersion != "", "the SARIF version has a value")
	return nil
}

// runScan loads the ruleset, scans the paths and writes the report.
func runScan(cmd *cobra.Command, f *scanFlags, paths []string) error {
	assert.That(cmd != nil && f != nil, "cmd and f are not nil")
	if err := f.validate(); err != nil {
		return err
	}
	rs, err := ruleset.Load(f.rules.options(nil))
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		paths = []string{"."}
	}
	exclude := make([]string, 0, len(rs.Exclude))
	for pattern := range rs.Exclude {
		exclude = append(exclude, pattern)
	}
	sort.Strings(exclude)
	eng, err := engine.New(rs.Rules, engine.DefaultLimits(), exclude...)
	if err != nil {
		return err
	}
	defer eng.Close()
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	res, err := eng.Scan(root, paths)
	if err != nil {
		return err
	}
	if err := writeReport(cmd.OutOrStdout(), f, res, rs.Active()); err != nil {
		return err
	}
	assert.That(res.Files >= 0, "the scan counts its files")
	if failing(res, f.failLevel) {
		return errFindings
	}
	return nil
}

// writeReport writes res in the selected format to the --output file or to stdout.
func writeReport(stdout io.Writer, f *scanFlags, res engine.Result, active []ruleset.Rule) (err error) {
	assert.That(stdout != nil && f != nil, "stdout and f are not nil")
	w := stdout
	if f.output != "" {
		file, cerr := os.Create(f.output)
		if cerr != nil {
			return cerr
		}
		defer func() {
			if cerr := file.Close(); err == nil {
				err = cerr
			}
		}()
		w = file
	}
	switch f.format {
	case "json":
		return report.WriteJSON(w, res, version.Get().Version)
	case "sarif":
		return report.WriteSARIF(w, f.sarifVersion, res, active, version.Get().Version)
	}
	assert.That(f.format == "text", "validate allows three formats")
	return report.WriteText(w, res)
}

// failing reports whether a finding has the level failLevel or a higher level.
func failing(res engine.Result, failLevel string) bool {
	rank := map[string]int{"note": 1, "warning": 2, "error": 3, "none": 4}
	limit, ok := rank[failLevel]
	assert.That(ok, "validate checks the fail level")
	assert.That(limit >= 1 && limit <= 4, "the rank is 1 to 4")
	for _, finding := range res.Findings {
		if rank[finding.Severity] >= limit {
			return true
		}
	}
	return false
}
