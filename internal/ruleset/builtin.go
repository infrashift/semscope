package ruleset

import (
	"sort"

	"github.com/infrashift/semscope/internal/assert"
	nasap10 "github.com/infrashift/semscope/rules/nasa_p10"
)

// Builtin returns the CUE source of the built-in ruleset name.
// A switch gives the sources, so the package has no mutable registry.
func Builtin(name string) ([]byte, bool) {
	switch name {
	case nasap10.Name:
		src := nasap10.Source()
		assert.That(len(src) > 0, "the nasa-p10 source is not empty")
		return src, true
	default:
		return nil, false
	}
}

// BuiltinNames returns the names of the built-in rulesets, sorted.
func BuiltinNames() []string {
	out := []string{nasap10.Name}
	assert.That(sort.StringsAreSorted(out), "the names are sorted")
	for _, name := range out {
		_, ok := Builtin(name)
		assert.That(ok, "each listed name has a built-in source")
	}
	return out
}
