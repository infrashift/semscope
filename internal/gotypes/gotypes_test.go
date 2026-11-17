package gotypes

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

const fixture = "../../testdata/go_types/drops.go"

func TestFindReportsEachDrop(t *testing.T) {
	abs, err := filepath.Abs(fixture)
	if err != nil {
		t.Fatal(err)
	}
	drops, err := Find([]string{abs}, 10)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, d := range drops {
		if d.File != abs {
			t.Errorf("drop in %s, want %s", d.File, abs)
		}
		got = append(got, fmt.Sprintf("%d:%d %s %s (%s)", d.Line, d.Col, d.Stmt, d.Callee, d.Arg))
	}
	want := []string{
		"32:2 call example.com/gotypes.fail ()",
		"33:2 call example.com/gotypes.pair ()",
		"34:8 defer (example.com/gotypes.closer).Close ()",
		"35:5 go example.com/gotypes.fail ()",
		"36:2 call example.com/gotypes.codedFail ()",
		"40:2 call (*strings.Builder).WriteString (string)",
		"41:2 call fmt.Fprintf (*strings.Builder)",
		"42:2 call fmt.Fprintln (*os.File)",
		"43:2 call os.Remove (string)",
		"45:2 call f ()",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("drops:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestIgnored(t *testing.T) {
	d := Drop{Callee: "fmt.Fprintf", Arg: "*strings.Builder", Stmt: StmtCall}
	if !d.Ignored([]string{"fmt.Fprintf(*strings.Builder)"}) || !d.Ignored([]string{"fmt.Fprintf"}) {
		t.Error("an entry with the name or with the name and the argument type ignores the drop")
	}
	if d.Ignored([]string{"fmt.Fprintf(*bytes.Buffer)", "fmt.Fprint"}) {
		t.Error("an entry for another argument type or another name ignores the drop")
	}
}

func TestFindErrors(t *testing.T) {
	dir := t.TempDir()
	if _, err := Find([]string{filepath.Join(dir, "missing.go")}, 10); err == nil {
		t.Error("a missing file gave no error")
	}
	if _, err := moduleRoot(dir); err == nil || !strings.Contains(err.Error(), "no go.mod") {
		t.Errorf("moduleRoot(%s) error = %v", dir, err)
	}
}
