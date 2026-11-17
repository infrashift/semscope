package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/infrashift/semscope/internal/assert"
	"github.com/infrashift/semscope/internal/ruleset"
	"github.com/infrashift/semscope/internal/version"
)

// scanFlags holds the flags of `semscope scan`.
type scanFlags struct {
	rules        rulesetFlags
	format       string
	sarifVersion string
	output       string
}

func newScanCmd() *cobra.Command {
	f := &scanFlags{}
	cmd := &cobra.Command{
		Use:   "scan [paths...]",
		Short: "Scan source files with the active rules",
		Long: `Scan the source files under each path with the active rules of the ruleset.
The default path is the current directory.
The exit status is 1 when a finding has the error level.`,
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
	if !version.IsSupportedSARIF(f.sarifVersion) {
		return fmt.Errorf("--sarif-version %q: want 2.1.0 or 2.2", f.sarifVersion)
	}
	assert.That(f.format != "", "the format has a value")
	assert.That(f.sarifVersion != "", "the SARIF version has a value")
	return nil
}

// runScan loads the ruleset for the scan. Task 4.1 adds the engine.
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
	assert.That(len(paths) > 0, "the scan has one or more paths")
	return fmt.Errorf("scan: %d active rule(s) for %d path(s): the scan engine arrives in Task 4.1",
		len(rs.Active()), len(paths))
}
