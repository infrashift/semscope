package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/infrashift/semscope/internal/assert"
	"github.com/infrashift/semscope/internal/version"
)

// The exit codes of semscope. The CLI contract in the wiki defines them.
const (
	exitOK       = 0
	exitFindings = 1
	exitUsage    = 2
)

// errFindings tells run that the scan wrote its report and found errors.
// semscope: allow(NASA-P10-06) a sentinel error value; the program never writes it
var errFindings = errors.New("findings at the error level")

// run executes the command line args and returns the process exit code.
// It prints each error once, on stderr.
func run(args []string, stdout, stderr io.Writer) int {
	assert.That(args != nil, "args is not nil, so cobra does not read os.Args")
	assert.That(stdout != nil && stderr != nil, "stdout and stderr are not nil")
	root := newRootCmd(stdout, stderr)
	root.SetArgs(args)
	err := root.Execute()
	switch {
	case err == nil:
		return exitOK
	case errors.Is(err, errFindings):
		return exitFindings
	}
	if _, werr := fmt.Fprintf(stderr, "semscope: %v\n", err); werr != nil {
		return exitUsage
	}
	return exitUsage
}

// newRootCmd builds the command tree. The commands write to stdout and stderr.
func newRootCmd(stdout, stderr io.Writer) *cobra.Command {
	info := version.Get()
	root := &cobra.Command{
		Use:   "semscope",
		Short: "Scan source code with composable CUE rulesets",
		Long: `semscope scans Go, C, C++, Zig, Python, Rust and C# source code with CUE rulesets.
The built-in ruleset nasa-p10 holds the NASA Power of 10 rules.
semscope writes the findings as text, JSON or SARIF.`,
		Version:       info.String(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetVersionTemplate("{{.Version}}")
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.CompletionOptions.DisableDefaultCmd = true
	root.AddCommand(newVersionCmd(info), newScanCmd(), newRulesCmd())
	assert.That(root.HasSubCommands(), "the root command has subcommands")
	assert.That(root.Version != "", "the root command has a version")
	return root
}

// newVersionCmd builds `semscope version`. It prints the same text as --version.
func newVersionCmd(info version.Info) *cobra.Command {
	assert.That(info.Version != "", "info comes from version.Get")
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version, the commit, the build date and the SARIF versions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := io.WriteString(cmd.OutOrStdout(), info.String())
			return err
		},
	}
	assert.That(cmd.Args != nil, "the version command takes no arguments")
	return cmd
}
