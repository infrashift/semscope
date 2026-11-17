// Package ruleset loads CUE rulesets, applies overlays and overrides, and
// gives the resolved rules.
//
// The loader unifies each source with the #Ruleset definition of
// schema.cue, in the sequence of the sources. It then applies the --set
// values, validates the result and decodes it. Last, it applies the
// --enable-rule and --omit-rule lists.
package ruleset

import (
	_ "embed"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueerrors "cuelang.org/go/cue/errors"

	"github.com/infrashift/semscope/internal/assert"
	nasap10 "github.com/infrashift/semscope/rules/nasa_p10"
)

//go:embed schema.cue
var schemaSource string

// Limits of the loader. Input that exceeds a limit gives an error.
const (
	MaxSources     = 64
	MaxSourceBytes = 1 << 20
	MaxRules       = 4096
	MaxParams      = 64
	MaxOverrides   = 256
	DefaultRuleset = nasap10.Name
)

// ParamKind is the type of a rule parameter.
type ParamKind int

// The parameter types that the schema allows.
const (
	ParamInt ParamKind = iota + 1
	ParamString
	ParamBool
	ParamList
)

// Param is one resolved rule parameter. Kind selects the field that holds the value.
type Param struct {
	Kind ParamKind
	Int  int64
	Str  string
	Bool bool
	List []string
}

// String gives the parameter value in CUE syntax.
func (p Param) String() string {
	assert.That(p.Kind >= ParamInt && p.Kind <= ParamList, "the parameter comes from decodeParam")
	switch p.Kind {
	case ParamInt:
		return strconv.FormatInt(p.Int, 10)
	case ParamString:
		return strconv.Quote(p.Str)
	case ParamBool:
		return strconv.FormatBool(p.Bool)
	case ParamList:
		quoted := make([]string, len(p.List))
		for i, s := range p.List {
			quoted[i] = strconv.Quote(s)
		}
		return "[" + strings.Join(quoted, ", ") + "]"
	}
	return ""
}

// Check is the analyzer configuration of a rule. Kind selects the checker.
// The other fields apply to some kinds only. Each map key is a language.
type Check struct {
	Kind             string              `json:"kind"`
	BannedCalls      map[string][]string `json:"banned_calls"`
	AllocCalls       map[string][]string `json:"alloc_calls"`
	InitFunctions    []string            `json:"init_functions"`
	AssertCalls      map[string][]string `json:"assert_calls"`
	BannedDirectives map[string][]string `json:"banned_directives"`
	MustCheckCalls   map[string][]string `json:"must_check_calls"`
	Query            map[string]string   `json:"query"`
	Patterns         map[string][]string `json:"patterns"`
}

// Rule is one resolved rule.
type Rule struct {
	ID        string           `json:"id"`
	Enabled   bool             `json:"enabled"`
	Severity  string           `json:"severity"`
	Languages []string         `json:"languages"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	Check     Check            `json:"check"`
	Params    map[string]Param `json:"-"`
}

// Resolved is the result of Load.
type Resolved struct {
	// Sources holds the name of each source, in load sequence.
	Sources []string
	// Rules holds every rule, sorted by ID. An omitted rule has Enabled false.
	Rules []Rule
}

// Options selects the sources and the overrides for Load.
type Options struct {
	// Rulesets holds built-in names, .cue files and directories. Empty means DefaultRuleset.
	Rulesets []string
	// Enable and Omit hold rule IDs. An ID in both lists gives an error.
	Enable []string
	Omit   []string
	// Set holds parameter overrides in the form RULE_ID.param=value.
	Set []string
}

// Active returns the enabled rules, sorted by ID.
func (r Resolved) Active() []Rule {
	assert.That(len(r.Rules) <= MaxRules, "the rule count is within MaxRules")
	out := make([]Rule, 0, len(r.Rules))
	for _, rule := range r.Rules {
		if rule.Enabled {
			out = append(out, rule)
		}
	}
	assert.That(len(out) <= len(r.Rules), "a filter gives no new rules")
	return out
}

// Omitted returns the disabled rules, sorted by ID.
func (r Resolved) Omitted() []Rule {
	assert.That(len(r.Rules) <= MaxRules, "the rule count is within MaxRules")
	out := make([]Rule, 0, len(r.Rules))
	for _, rule := range r.Rules {
		if !rule.Enabled {
			out = append(out, rule)
		}
	}
	assert.That(len(out) <= len(r.Rules), "a filter gives no new rules")
	return out
}

// Load reads the sources of opts, applies the overrides and returns the rules.
func Load(opts Options) (Resolved, error) {
	switch {
	case len(opts.Rulesets) > MaxSources:
		return Resolved{}, fmt.Errorf("ruleset: %d rulesets given (max %d)", len(opts.Rulesets), MaxSources)
	case len(opts.Set) > MaxOverrides:
		return Resolved{}, fmt.Errorf("ruleset: %d --set values given (max %d)", len(opts.Set), MaxOverrides)
	case len(opts.Enable) > MaxRules || len(opts.Omit) > MaxRules:
		return Resolved{}, fmt.Errorf("ruleset: more than %d rule IDs in --enable-rule or --omit-rule", MaxRules)
	}
	ctx := cuecontext.New()
	v, sources, err := unifySources(ctx, opts.Rulesets)
	if err != nil {
		return Resolved{}, err
	}
	if v, err = applySets(v, opts.Set); err != nil {
		return Resolved{}, err
	}
	if err = v.Validate(cue.Final(), cue.Concrete(true)); err != nil {
		return Resolved{}, cueError("invalid ruleset", err)
	}
	rules, err := decodeRules(v)
	if err != nil {
		return Resolved{}, err
	}
	if err = applyToggles(rules, opts.Enable, opts.Omit); err != nil {
		return Resolved{}, err
	}
	assert.That(len(sources) > 0, "Load reads one or more sources")
	assert.That(len(rules) <= MaxRules, "decodeRules enforces MaxRules")
	assert.That(sort.SliceIsSorted(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID }),
		"the rules are sorted by ID")
	return Resolved{Sources: sources, Rules: rules}, nil
}

// cueError formats a CUE error with its source positions.
func cueError(context string, err error) error {
	assert.That(err != nil, "cueError needs an error")
	details := strings.TrimSpace(cueerrors.Details(err, nil))
	assert.That(details != "", "a CUE error has details")
	return fmt.Errorf("ruleset: %s: %s", context, details)
}
