// Package gotypes is the fixture of the type-aware Go check of NASA-P10-07.
package gotypes

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type closer struct{}

func (closer) Close() error { return nil }

type coded interface {
	error
	Code() int
}

func fail() error { return errors.New("fail") }

func pair() (int, error) { return 0, nil }

func codedFail() coded { return nil }

func none() {}

func value() int { return 1 }

// Use holds one drop of each kind and the calls that are not drops.
func Use() error {
	fail()
	pair()
	defer closer{}.Close()
	go fail()
	codedFail()
	none()
	value()
	var b strings.Builder
	b.WriteString("ignored")
	fmt.Fprintf(&b, "ignored")
	fmt.Fprintln(os.Stderr, "not ignored")
	os.Remove("x") // semscope: allow(NASA-P10-07) the fixture tests the suppression
	f := fail
	f()
	if err := fail(); err != nil {
		return err
	}
	return fail()
}
