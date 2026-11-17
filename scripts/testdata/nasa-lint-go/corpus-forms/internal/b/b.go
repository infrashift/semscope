// Package b holds one-line functions and the R9 forms of cobra and pflag.
package b

import "sort"

type byName []string

func (s byName) Len() int           { return len(s) }
func (s byName) Less(i, j int) bool { return s[i] < s[j] }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// Loop recurses in one line. R1.
func Loop(n int) int { return Loop(n - 1) }

// Empty has a struct literal before the body brace.
func Empty() struct{} { return struct{}{} }

// Alias is a function type alias. R9.
type Alias = func(int) error

// OnInit takes variadic funcs. R9.
func OnInit(y ...func()) {
	sort.Sort(byName(nil))
	run(y)
}

func run(y []func()) {
	for range y {
	}
}

// Short is a one-line function. The next brace line is not its body.
func Short() int { return 1 }

var table = map[string]int{
	"a": 1,
}
