package a

import "example.com/fx/internal/assert"

func foo() (int, error) {
	assert.That(true, "x")
	assert.That(1 > 0, "x")
	return 1, nil
}

func drops(m map[string]int, x interface{}) error {
	assert.That(m != nil, "m")
	assert.That(x != nil, "x")
	_ = bar()
	v, _ := foo()
	v, _ = foo()
	_, err := foo()
	w, _ := m["k"]
	n, _ := x.(int)
	_ = v + w + n
	return err
}

func bar() error {
	assert.That(true, "x")
	assert.That(1 > 0, "x")
	var y, _ = foo()
	return nil
}
