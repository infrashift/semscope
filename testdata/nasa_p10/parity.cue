// Overlay for the parity tests (ADR-0005). It keeps only the checks that
// scripts/nasa-lint.py and scripts/nasa-lint-go.py also make.
package overlay

rules: {
	// The oracles do not check R3, R6, R8 or R10.
	"NASA-P10-03": enabled: false
	"NASA-P10-06": enabled: false
	"NASA-P10-08": enabled: false
	"NASA-P10-10": enabled: false
	// The oracles check function pointers, but not the pointer depth.
	"NASA-P10-09": params: max_pointer_depth: 1000
}
