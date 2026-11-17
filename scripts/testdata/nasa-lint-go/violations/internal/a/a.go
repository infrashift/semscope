// Package a holds fixture violations for nasa-lint-go.
package a

import (
	"errors"
	"sort"
	"strconv"
)

// Handler is a function type. R9.
type Handler func(int) error

// T is a fixture type.
type T struct {
	cb    func()
	table map[string]func()
	items []int
}

var hook func()

// Fact recurses directly. R1.
func Fact(n int) int {
	if n <= 1 {
		return 1
	}
	return n * Fact(n-1)
}

// walk recurses through a method of the receiver type. R1.
func (t *T) walk(n int) {
	t.step(n)
}

func (t *T) step(n int) {
	if n > 0 {
		t.walk(n - 1)
	}
}

// Loop has an unbounded loop. R2.
func Loop(ch chan int) {
	for {
		if <-ch == 0 {
			return
		}
	}
}

// Spin has the other unbounded loop forms. R2.
func Spin(ch chan int) {
	for ;; {
		<-ch
	}
	for true {
		<-ch
	}
}

// Serve has a suppressed unbounded loop.
func Serve(ch chan int) {
	// nasa-lint: allow(R2) the loop ends when ch closes
	for {
		if _, ok := <-ch; !ok {
			return
		}
	}
}

// Drop discards call results. R7 on the first two lines only.
func Drop() error {
	_ = errors.New("x")
	v, _ := strconv.Atoi("1")
	_, err := strconv.Atoi("2")
	m := map[string]int{}
	w, _ := m["k"]
	var i interface{} = 1
	n, _ := i.(int)
	_ = v + w + n
	return err
}

// Jump uses goto. R1.
func Jump(n int) int {
start:
	if n > 0 {
		n--
		goto start
	}
	return n
}

// Text holds rule patterns in literals and comments: for { goto _ = f(
func Text() string {
	s := "for { goto x; _ = f()"
	r := `
_ = f()
for {
`
	c := '{'
	/* for {
	goto x */
	return s + r + string(c)
}

// Apply takes a func-typed parameter. R9.
func Apply(n int, fn func(int) int) int {
	return fn(n)
}

// Make returns a func type. R9.
func Make() func() int {
	return func() int { return 1 }
}

// Run uses closures that are not findings.
func Run(xs []int) {
	go func() {}()
	defer func() {}()
	sort.Slice(xs, func(i, j int) bool { return xs[i] < xs[j] })
}

// testRecurse is a test helper: exempt from R1, R4 and R5.
func testRecurse(n int) int {
	return testRecurse(n)
}

// Long has 61 lines. R4.
func Long() int {
	x := 0
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	return x
}

// Short has 60 lines. Not a finding.
func Short() int {
	x := 0
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	x++
	return x
}
