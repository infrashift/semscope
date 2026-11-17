package main

import (
	"github.com/spf13/pflag"

	"github.com/infrashift/semscope/internal/assert"
	"github.com/infrashift/semscope/internal/ruleset"
)

// rulesetFlags holds the flags that select and change a ruleset.
type rulesetFlags struct {
	rulesets []string
	enable   []string
	omit     []string
	disable  []string
	set      []string
}

// bind adds the flags to fs. withRuleset adds --ruleset and -r.
func (f *rulesetFlags) bind(fs *pflag.FlagSet, withRuleset bool) {
	assert.That(fs != nil, "fs is not nil")
	if withRuleset {
		fs.StringArrayVarP(&f.rulesets, "ruleset", "r", nil,
			"a built-in ruleset, a .cue file or a directory; repeat the flag to add overlays (default nasa-p10)")
	}
	fs.StringSliceVar(&f.enable, "enable-rule", nil, "enable the rules with these IDs")
	fs.StringSliceVar(&f.omit, "omit-rule", nil, "omit the rules with these IDs")
	fs.StringSliceVar(&f.disable, "disable-rule", nil, "the same as --omit-rule")
	fs.StringArrayVar(&f.set, "set", nil, "override a rule parameter: RULE_ID.param=value (a list value is comma-separated)")
	assert.That(fs.Lookup("set") != nil, "the --set flag exists")
}

// options gives the ruleset.Options for the flags. Each extra value goes
// after the --ruleset values.
func (f *rulesetFlags) options(extra []string) ruleset.Options {
	rulesets := append(append([]string{}, f.rulesets...), extra...)
	omit := append(append([]string{}, f.omit...), f.disable...)
	assert.That(len(rulesets) == len(f.rulesets)+len(extra), "no ruleset value is lost")
	assert.That(len(omit) == len(f.omit)+len(f.disable), "no omitted rule is lost")
	return ruleset.Options{Rulesets: rulesets, Enable: f.enable, Omit: omit, Set: f.set}
}
