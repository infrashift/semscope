package parser

import (
	"path/filepath"
	"strings"

	ts "github.com/infrashift/ztree/bindings/go"

	"github.com/infrashift/semscope/internal/assert"
)

// Language names a language that the parser reads. The values match the
// #Language values of the ruleset schema.
type Language string

// The languages.
const (
	Go     Language = "go"
	C      Language = "c"
	CPP    Language = "cpp"
	Zig    Language = "zig"
	Python Language = "python"
	Rust   Language = "rust"
	CSharp Language = "csharp"
)

// Languages returns all languages in a fixed sequence.
func Languages() []Language {
	out := []Language{Go, C, CPP, Zig, Python, Rust, CSharp}
	assert.That(len(out) == 7, "seven languages")
	return out
}

// Detect returns the language of path from its file extension.
// A .h file counts as C.
func Detect(path string) (Language, bool) {
	assert.That(path != "", "the path is not empty")
	var lang Language
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		lang = Go
	case ".c", ".h":
		lang = C
	case ".cc", ".cpp", ".cxx", ".hh", ".hpp", ".hxx":
		lang = CPP
	case ".zig":
		lang = Zig
	case ".py":
		lang = Python
	case ".rs":
		lang = Rust
	case ".cs":
		lang = CSharp
	}
	_, known := Grammar(lang)
	assert.That(lang == "" || known, "each detected language has a grammar")
	return lang, lang != ""
}

// Grammar returns the Tree-sitter grammar of lang.
// An empty or unknown lang gives false.
func Grammar(lang Language) (*ts.Language, bool) {
	var g *ts.Language
	switch lang {
	case Go:
		g = ts.NewLanguage(ts.LangGo)
	case C:
		g = ts.NewLanguage(ts.LangC)
	case CPP:
		g = ts.NewLanguage(ts.LangCPP)
	case Zig:
		g = ts.NewLanguage(ts.LangZig)
	case Python:
		g = ts.NewLanguage(ts.LangPython)
	case Rust:
		g = ts.NewLanguage(ts.LangRust)
	case CSharp:
		g = ts.NewLanguage(ts.LangCSharp)
	}
	assert.That(g == nil || g.AbiVersion() > 0, "a grammar has an ABI version")
	return g, g != nil
}

// isTestFile reports whether the whole file holds tests: a Go _test.go file,
// a Python test_*.py or *_test.py file, or a C# *Test.cs or *Tests.cs file.
func isTestFile(path string, lang Language) bool {
	assert.That(path != "", "the path is not empty")
	assert.That(lang != "", "the language is not empty")
	base := filepath.Base(path)
	assert.That(base != "" && base != "/", "a file has a base name")
	switch lang {
	case Go:
		return strings.HasSuffix(base, "_test.go")
	case Python:
		return strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.py")
	case CSharp:
		return strings.HasSuffix(base, "Test.cs") || strings.HasSuffix(base, "Tests.cs")
	}
	return false
}

// isTestHelper reports whether a function name marks a test function or a
// test helper. Such a name starts with "test" or "Test", or ends with "Test".
func isTestHelper(name string) bool {
	assert.That(name != "", "a function has a name")
	return strings.HasPrefix(name, "test") || strings.HasPrefix(name, "Test") || strings.HasSuffix(name, "Test")
}
