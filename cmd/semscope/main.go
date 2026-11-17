// Command semscope scans source code with composable CUE rulesets and reports
// the findings as SARIF, JSON or text.
package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

// run executes the command line args and returns the process exit code.
// Task 2.2 replaces this placeholder with the cobra command tree.
func run(args []string, stderr io.Writer) int {
	if _, err := fmt.Fprintf(stderr, "semscope: %d argument(s) received; the commands arrive in Task 2.2\n", len(args)); err != nil {
		return 1
	}
	return 2
}
