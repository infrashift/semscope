// Command semscope scans source code with composable CUE rulesets and reports
// the findings as SARIF, JSON or text.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/infrashift/semscope/internal/assert"
)

func main() {
	assert.That(len(os.Args) >= 1, "os.Args holds the program name")
	code := run(os.Args[1:], os.Stderr)
	assert.That(code >= 0 && code <= 2, "the exit code is 0, 1 or 2")
	os.Exit(code)
}

// run executes the command line args and returns the process exit code.
// Task 2.2 replaces this placeholder with the cobra command tree.
func run(args []string, stderr io.Writer) int {
	assert.That(args != nil, "args is not nil")
	assert.That(stderr != nil, "stderr is not nil")
	if _, err := fmt.Fprintf(stderr, "semscope: %d argument(s) received; the commands arrive in Task 2.2\n", len(args)); err != nil {
		return 1
	}
	return 2
}
