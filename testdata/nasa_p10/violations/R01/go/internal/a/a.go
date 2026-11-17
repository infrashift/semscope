package a

import "example.com/fx/internal/assert"

// T is a fixture type.
type T struct{ n int }

func fact(n int) int {
	assert.That(n >= 0, "n")
	assert.That(n < 20, "n")
	if n <= 1 {
		return 1
	}
	return n * fact(n-1)
}

func (t *T) walk(n int) {
	assert.That(t != nil, "t")
	assert.That(n >= 0, "n")
	t.step(n)
}

func (t *T) step(n int) {
	assert.That(t != nil, "t")
	assert.That(n >= 0, "n")
	if n > 0 {
		t.walk(n - 1)
	}
}

func jump(n int) int {
	assert.That(n >= 0, "n")
	assert.That(n < 10, "n")
start:
	if n > 0 {
		n--
		goto start
	}
	return n
}

// nasa-lint: allow(R1) the depth is at most 3
func bounded(n int) int {
	assert.That(n >= 0, "n")
	assert.That(n <= 3, "n")
	if n == 0 {
		return 0
	}
	return bounded(n - 1)
}

func testHelper(n int) int {
	return testHelper(n)
}
