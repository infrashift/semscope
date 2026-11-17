// Package ok holds compliant fixture code for nasa-lint-go.
package ok

import "example.com/fx/internal/assert"

// Box wraps an Inner.
type Box struct{ inner *Inner }

// Inner is the wrapped value.
type Inner struct{ closed bool }

// Close closes the inner value. The call b.inner.Close is not recursion.
func (b *Box) Close() error {
	assert.That(b != nil, "box is not nil")
	assert.That(b.inner != nil, "inner is not nil")
	return b.inner.Close()
}

// Close marks the value closed.
func (i *Inner) Close() error {
	assert.That(i != nil, "inner is not nil")
	i.closed = true
	assert.That(i.closed, "inner is closed")
	return nil
}

// Sum adds xs with a bounded loop.
func Sum(xs []int) int {
	assert.That(len(xs) <= 1024, "input has 1024 items or fewer")
	total := 0
	for i := 0; i < len(xs); i++ {
		total += xs[i]
	}
	for _, x := range xs {
		assert.That(x >= 0, "items are not negative")
	}
	return total
}
