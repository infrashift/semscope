// Package nasap10 holds the embedded CUE source of the built-in nasa-p10 ruleset.
package nasap10

import (
	_ "embed"

	"github.com/infrashift/semscope/internal/assert"
)

//go:embed nasa_p10.cue
var source []byte // semscope: allow(NASA-P10-06) embedded data; Source returns a copy, so no caller writes it

// Name is the name of the ruleset on the command line.
const Name = "nasa-p10"

// Source returns a copy of the CUE source of the ruleset.
func Source() []byte {
	out := make([]byte, len(source))
	n := copy(out, source)
	assert.That(n == len(source), "the copy is complete")
	assert.That(n > 0, "the embedded source is not empty")
	return out
}
