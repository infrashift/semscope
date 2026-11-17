package main

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/infrashift/semscope/internal/assert"
	"github.com/infrashift/semscope/internal/ruleset"
)

func newRulesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "List and inspect rulesets",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newRulesListCmd(), newRulesInspectCmd())
	assert.That(len(cmd.Commands()) == 2, "rules has two subcommands")
	assert.That(cmd.Args != nil, "rules takes no arguments")
	return cmd
}

func newRulesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the built-in rulesets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return writeRulesList(cmd.OutOrStdout())
		},
	}
	assert.That(cmd.RunE != nil, "the list command runs")
	assert.That(cmd.Args != nil, "list takes no arguments")
	return cmd
}

// writeRulesList writes one line for each built-in ruleset.
func writeRulesList(w io.Writer) error {
	assert.That(w != nil, "w is not nil")
	assert.That(len(ruleset.BuiltinNames()) > 0, "semscope has one or more built-in rulesets")
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, name := range ruleset.BuiltinNames() {
		rs, err := ruleset.Load(ruleset.Options{Rulesets: []string{name}})
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(tw, "%s\t%d rules\t%s\n", name, len(rs.Rules), rs.Description); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func newRulesInspectCmd() *cobra.Command {
	f := &rulesetFlags{}
	cmd := &cobra.Command{
		Use:   "inspect <ruleset> [overlay...]",
		Short: "Show the active rules, the omitted rules and the effective parameters",
		Long: `Load a ruleset and the overlays in sequence.
Apply the flags.
Show the active rules, the omitted rules and the effective parameters.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rs, err := ruleset.Load(f.options(args))
			if err != nil {
				return err
			}
			return writeInspect(cmd.OutOrStdout(), rs)
		},
	}
	f.bind(cmd.Flags(), false)
	assert.That(cmd.Flags().Lookup("ruleset") == nil, "inspect takes the rulesets as arguments")
	assert.That(cmd.Args != nil, "inspect needs one or more arguments")
	return cmd
}

// writeInspect writes the sources, the active rules and the omitted rules of rs.
func writeInspect(w io.Writer, rs ruleset.Resolved) error {
	assert.That(w != nil, "w is not nil")
	var b strings.Builder
	fmt.Fprintf(&b, "ruleset: %s\nsources: %s\n", rs.Name, strings.Join(rs.Sources, ", "))
	patterns := make([]string, 0, len(rs.Exclude))
	for p := range rs.Exclude {
		patterns = append(patterns, p)
	}
	sort.Strings(patterns)
	for _, p := range patterns {
		fmt.Fprintf(&b, "exclude: %s (%s)\n", p, rs.Exclude[p])
	}
	active, omitted := rs.Active(), rs.Omitted()
	fmt.Fprintf(&b, "\nactive rules: %d\n", len(active))
	writeRuleTable(&b, active)
	fmt.Fprintf(&b, "\nomitted rules: %d\n", len(omitted))
	writeRuleTable(&b, omitted)
	assert.That(len(active)+len(omitted) == len(rs.Rules), "each rule is active or omitted")
	_, err := io.WriteString(w, b.String())
	return err
}

// writeRuleTable writes one block for each rule: ID, severity, kind, title,
// languages and parameters. A strings.Builder does not fail, so the writes
// have no error to check.
func writeRuleTable(b *strings.Builder, rules []ruleset.Rule) {
	assert.That(b != nil, "b is not nil")
	for _, r := range rules {
		fmt.Fprintf(b, "  %s  %s  %s  %s\n", r.ID, r.Severity, r.Check.Kind, r.Title)
		fmt.Fprintf(b, "      languages: %s\n", strings.Join(r.Languages, ", "))
		if len(r.Params) == 0 {
			continue
		}
		names := make([]string, 0, len(r.Params))
		for name := range r.Params {
			names = append(names, name)
		}
		sort.Strings(names)
		parts := make([]string, len(names))
		for i, name := range names {
			parts[i] = name + "=" + r.Params[name].String()
		}
		fmt.Fprintf(b, "      params: %s\n", strings.Join(parts, " "))
	}
	assert.That(b.Len() > 0 || len(rules) == 0, "each rule writes one or more lines")
}
