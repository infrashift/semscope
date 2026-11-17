package assert

// That panics when cond is false. The primitive is exempt from R5.
func That(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}
