package a

import "example.com/fx/internal/assert"

func one(n int) int {
	assert.That(n < 100, "n")
	return n
}

func two(n int) int {
	assert.That(n < 100, "n")
	assert.That(n >= 0, "n")
	return n
}

func none(n int) int {
	return n
}
