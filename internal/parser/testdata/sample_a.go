package a

import "unsafe"

var hook func()
var count = 0

type H func(int) error

//go:linkname x y
func (t *T) walk(n int) (int, error) {
	assert.That(n > 0, "x")
	for {
	}
	for i := 0; i < n; i++ {
	}
	_ = foo()
	v, _ := bar()
	goto end
end:
	t.step(n)
	var pp **int
	return 0, nil
}

func helper(fn func(int)) func() { return nil }

// nolint
func x() {}
