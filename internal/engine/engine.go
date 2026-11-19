// Package engine runs the checks of the resolved rules on the parsed files,
// with fixed bounds on all work.
//
// New compiles the patterns and queries of the rules and allocates a fixed
// pool of workers. Scan collects the files and parses each file in the pool.
// It runs the per-file checks, then the checks over all files: recursion and
// the assertion average. Scan sorts the findings, so the output is the same
// on each run.
package engine

import (
	"fmt"
	"path"
	"runtime"

	ts "github.com/infrashift/ztree/bindings/go"

	"github.com/infrashift/semscope/internal/assert"
	"github.com/infrashift/semscope/internal/parser"
	"github.com/infrashift/semscope/internal/ruleset"
)

// Limits bound the work of one scan. Work that exceeds a limit gives an error.
type Limits struct {
	Parser        parser.Limits
	MaxFiles      int
	MaxFindings   int
	MaxGraphNodes int
	Workers       int
}

// DefaultLimits returns the limits of the semscope command.
func DefaultLimits() Limits {
	l := Limits{
		Parser:        parser.DefaultLimits(),
		MaxFiles:      100_000,
		MaxFindings:   1_000_000,
		MaxGraphNodes: 1_000_000,
		Workers:       min(runtime.NumCPU(), 16),
	}
	assert.That(l.Workers >= 1, "one or more workers")
	assert.That(l.MaxFiles > 0 && l.MaxFindings > 0 && l.MaxGraphNodes > 0, "positive limits")
	return l
}

// Finding is one rule violation. A finding without a location has an empty Path.
type Finding struct {
	RuleID   string
	Severity string
	Message  string
	Path     string
	Line     int
	Col      int
	EndLine  int
	EndCol   int
}

// LangStats counts the functions and assertions of one language for the assertion average.
type LangStats struct {
	Functions  int
	Assertions int
}

// Result is the outcome of one scan.
type Result struct {
	Findings []Finding
	Files    int
	Stats    map[parser.Language]LangStats
}

// worker is one pre-allocated parse slot of the pool.
type worker struct {
	parser *parser.Parser
	cursor *ts.QueryCursor
}

// Engine runs the active rules. It is not safe for concurrent calls of Scan.
type Engine struct {
	rules   []compiledRule
	exclude []string
	langs   map[parser.Language]bool
	limits  Limits
	workers []worker
}

// New compiles the active rules and allocates the worker pool. A scan skips
// each path that an exclude pattern matches (see excluded).
func New(rules []ruleset.Rule, limits Limits, exclude ...string) (*Engine, error) {
	if limits.Workers < 1 || limits.MaxFiles < 1 || limits.MaxFindings < 1 || limits.MaxGraphNodes < 1 {
		return nil, fmt.Errorf("engine: each limit must be positive: %+v", limits)
	}
	for _, p := range exclude {
		if _, err := path.Match(p, ""); err != nil {
			return nil, fmt.Errorf("engine: exclude pattern %q: %w", p, err)
		}
	}
	e := &Engine{langs: map[parser.Language]bool{}, limits: limits, exclude: exclude}
	for _, r := range rules {
		if !r.Enabled {
			continue
		}
		cr, err := compileRule(r)
		if err != nil {
			e.Close()
			return nil, err
		}
		e.rules = append(e.rules, cr)
		for lang := range cr.langs {
			e.langs[lang] = true
		}
	}
	for i := 0; i < limits.Workers; i++ {
		p, err := parser.New(limits.Parser)
		if err != nil {
			e.Close()
			return nil, err
		}
		e.workers = append(e.workers, worker{parser: p, cursor: ts.NewQueryCursor()})
	}
	assert.That(len(e.workers) == limits.Workers, "the pool has every worker")
	assert.That(len(e.rules) <= len(rules), "the engine compiles the active rules only")
	return e, nil
}

// Close frees the compiled queries and the worker pool.
func (e *Engine) Close() {
	assert.That(e != nil, "e is not nil")
	for i := range e.rules {
		e.rules[i].close()
	}
	for _, w := range e.workers {
		w.parser.Close()
	}
	e.rules, e.workers = nil, nil
	assert.That(len(e.workers) == 0, "the pool is empty after Close")
}
