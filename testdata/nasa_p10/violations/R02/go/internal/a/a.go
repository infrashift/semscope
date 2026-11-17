package a

import "example.com/fx/internal/assert"

func spin(ch chan int) {
	assert.That(ch != nil, "ch")
	assert.That(cap(ch) >= 0, "cap")
	for {
		<-ch
	}
}

func spin2(ch chan int) {
	assert.That(ch != nil, "ch")
	assert.That(cap(ch) >= 0, "cap")
	for ;; {
		<-ch
	}
}

func spin3(ch chan int) {
	assert.That(ch != nil, "ch")
	assert.That(cap(ch) >= 0, "cap")
	for true {
		<-ch
	}
}

func served(ch chan int) {
	assert.That(ch != nil, "ch")
	assert.That(cap(ch) >= 0, "cap")
	// nasa-lint: allow(R2) the loop ends when ch closes
	for {
		<-ch
	}
}

func counted(n int) int {
	assert.That(n >= 0, "n")
	assert.That(n < 100, "n")
	s := "for {"
	for i := 0; i < n; i++ {
		s += "x"
	}
	return len(s)
}
