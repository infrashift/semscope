package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunPlaceholderReturnsUsageCode(t *testing.T) {
	var stderr bytes.Buffer
	if code := run([]string{"scan"}, &stderr); code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "1 argument(s)") {
		t.Fatalf("stderr = %q, want the argument count", stderr.String())
	}
}
