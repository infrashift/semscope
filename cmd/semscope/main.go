// Command semscope scans source code with composable CUE rulesets and reports
// the findings as SARIF, JSON or text.
package main

import (
	"os"

	"github.com/infrashift/semscope/internal/assert"
)

func main() {
	assert.That(len(os.Args) >= 1, "os.Args holds the program name")
	code := run(os.Args[1:], os.Stdout, os.Stderr)
	assert.That(code >= exitOK && code <= exitUsage, "the exit code is 0, 1 or 2")
	os.Exit(code)
}
