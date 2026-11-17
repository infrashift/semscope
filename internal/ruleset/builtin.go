package ruleset

import (
	"sort"

	"github.com/infrashift/semscope/internal/assert"
)

// Builtin returns the CUE source of the built-in ruleset name.
// A switch gives the sources, so the package has no mutable registry.
func Builtin(name string) ([]byte, bool) {
	switch name {
	case DefaultRuleset:
		// Task 3.2 adds the embedded nasa-p10 source.
		return nil, false
	default:
		return nil, false
	}
}

// BuiltinNames returns the names of the built-in rulesets, sorted.
func BuiltinNames() []string {
	out := []string{}
	assert.That(sort.StringsAreSorted(out), "the names are sorted")
	for _, name := range out {
		_, ok := Builtin(name)
		assert.That(ok, "each listed name has a built-in source")
	}
	return out
}
