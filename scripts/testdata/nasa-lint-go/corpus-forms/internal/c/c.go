// Package c holds the forms of the CUE and x/tools corpora (Task 6.3).
package c

import (
	"context"
	"errors"
	"fmt"
)

type list []error

// Msg calls Msg of an element, not of the receiver. No recursion.
func (p list) Msg() string {
	if len(p) == 0 {
		return ""
	}
	return p[0].Error()
}

// Error calls Error of an element, not of the receiver. No recursion.
func (p list) Error() string { return p[0].Error() }

type server struct{}

func (*server) Error() string { return "" }

// The conversion is not a call, so it is not a discarded result.
var _ error = (*server)(nil)

// Assert drops the ok of a type assertion, not a call result.
func Assert(ctx context.Context) string {
	s, _ := ctx.Value("k").(string)
	if e, _ := ctx.Value("e").(error); e != nil {
		return e.Error()
	}
	return s
}

// Chain drops the result of the last call of a chain. R7.
func Chain(ctx context.Context) {
	v, _ := context.WithCancel(ctx)
	_ = v
	_, _ = errors.New("x").Error(), 0
}

// Count has a loop without a condition. R2.
func Count(ch chan int) int {
	for i := 0; ; i++ {
		if <-ch == 0 {
			return i
		}
	}
}

// Start returns a func in a result list. R9.
func Start(ctx context.Context) (context.Context, func()) {
	return context.WithCancel(ctx)
}

// Hooks makes a slice of funcs. R9.
func Hooks(n int) int {
	hooks := make([]func(), n)
	return len(hooks)
}

// Apply calls a call result with a func literal. No R9.
func Apply() {
	apply(1)(func() {})
}

func apply(int) func(func()) { return func(func()) {} }

// Abs calls a method of a call result. No recursion.
func Abs(p string) bool {
	return getOS().Abs(p)
}

type osT struct{}

func (osT) Abs(string) bool { return true }

func getOS() osT { return osT{} }

// Deadline holds an interface method of the same name. No recursion.
func Deadline(v any) bool {
	_, ok := v.(interface {
		Deadline() (int, bool)
	})
	return ok
}

// Close drops an error in a deferred func literal. R7.
func Close(c interface{ Close() error }) {
	defer func() { _ = c.Close() }()
}

// Bisect has a loop with a post statement and no condition. R2.
func Bisect(ch chan int) {
	i := 0
	for ; ; i++ {
		<-ch
	}
}

// Fn asserts a func type. R9.
func Fn(v any) bool {
	_, ok := v.(func())
	return ok
}

// PreempterFunc is a function type. R9.
type PreempterFunc func()

// The typed var with a conversion checks the interface at compile time.
var _ fmt.Stringer = PreempterFunc(nil)

func (PreempterFunc) String() string { return "" }
