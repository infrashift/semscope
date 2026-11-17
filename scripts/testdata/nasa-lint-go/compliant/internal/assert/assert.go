// Package assert defines the assertion primitive. It is exempt from R5.
package assert

// That panics when cond is false.
func That(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}
