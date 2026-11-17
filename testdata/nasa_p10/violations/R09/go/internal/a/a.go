package a

import "example.com/fx/internal/assert"

// Handler is a function type.
type Handler func(int) error

// Ops holds function fields.
type Ops struct {
	run   func()
	table map[string]func()
	list  []func()
}

var hook func()

func apply(
	fn func(int) int,
	n int,
) int {
	assert.That(fn != nil, "fn")
	assert.That(n >= 0, "n")
	return fn(n)
}

func make2() func() int {
	assert.That(true, "x")
	assert.That(1 > 0, "x")
	return nil
}

func closures(xs []int) {
	assert.That(xs != nil, "xs")
	assert.That(len(xs) >= 0, "xs")
	go func() {}()
	defer func() {}()
	sortSlice(xs, func(i, j int) bool { return xs[i] < xs[j] })
	var pp **int
	_ = pp
}
