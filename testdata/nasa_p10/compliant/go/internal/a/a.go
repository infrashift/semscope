package a

import "example.com/fx/internal/assert"

// Sum adds xs with a bounded loop.
func Sum(xs []int) (int, error) {
	assert.That(len(xs) <= 1024, "xs")
	assert.That(xs != nil, "xs")
	total := 0
	for i := 0; i < len(xs); i++ {
		total += xs[i]
	}
	v, err := parse(total)
	if err != nil {
		return 0, err
	}
	s := "for { goto x; _ = f()"
	return v + len(s), nil
}

func parse(n int) (int, error) {
	assert.That(n >= 0, "n")
	assert.That(n < 1<<30, "n")
	return n, nil
}
