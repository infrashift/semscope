package parser

import (
	"bytes"
	"fmt"

	ts "github.com/infrashift/ztree/bindings/go"

	"github.com/infrashift/semscope/internal/assert"
)

// Limits bound the work for one file. Input that exceeds a limit gives an error.
type Limits struct {
	MaxFileBytes int
	MaxNodes     int
	MaxDepth     int
	MaxNesting   int
	MaxRecords   int
}

// DefaultLimits returns the limits of the semscope command.
func DefaultLimits() Limits {
	l := Limits{MaxFileBytes: 16 << 20, MaxNodes: 8_000_000, MaxDepth: 4096, MaxNesting: 256, MaxRecords: 2_000_000}
	assert.That(l.MaxFileBytes > 0 && l.MaxNodes > 0, "positive limits")
	assert.That(l.MaxDepth > 0 && l.MaxNesting > 0 && l.MaxRecords > 0, "positive limits")
	return l
}

// Parser parses files of all seven languages. It is not safe for concurrent
// use: give each worker its own Parser.
type Parser struct {
	ts       *ts.Parser
	grammars map[Language]*ts.Language
	current  Language
	limits   Limits
	w        walker
}

// New returns a Parser with the grammars of all languages.
func New(limits Limits) (*Parser, error) {
	if limits.MaxFileBytes <= 0 || limits.MaxNodes <= 0 || limits.MaxDepth <= 0 ||
		limits.MaxNesting <= 0 || limits.MaxRecords <= 0 {
		return nil, fmt.Errorf("parser: each limit must be positive: %+v", limits)
	}
	p := &Parser{ts: ts.NewParser(), grammars: make(map[Language]*ts.Language, 7), limits: limits}
	for _, lang := range Languages() {
		g, ok := Grammar(lang)
		assert.That(ok, "each language has a grammar")
		p.grammars[lang] = g
	}
	p.w.fnStack = make([]int, 0, limits.MaxNesting)
	p.w.fnEnds = make([]uint, 0, limits.MaxNesting)
	assert.That(len(p.grammars) == len(Languages()), "the parser holds every grammar")
	return p, nil
}

// Close releases the Tree-sitter parser. Go owns the memory of ztree.
func (p *Parser) Close() {
	assert.That(p.grammars != nil, "New made the grammars")
	assert.That(p.ts != nil, "the parser is open")
	p.ts = nil
	assert.That(p.ts == nil, "the parser is closed")
}

// Parse parses src as lang. For a test file it returns a File with Exempt
// set and a nil tree.
func (p *Parser) Parse(path string, lang Language, src []byte) (*File, *ts.Tree, error) {
	assert.That(p.ts != nil, "the parser is open")
	if len(src) > p.limits.MaxFileBytes {
		return nil, nil, fmt.Errorf("%s: %d bytes (max %d)", path, len(src), p.limits.MaxFileBytes)
	}
	f := &File{Path: path, Lang: lang, Lines: splitLines(src), Imports: map[string]string{}}
	if isTestFile(path, lang) {
		f.Exempt = true
		return f, nil, nil
	}
	if p.current != lang {
		g, ok := p.grammars[lang]
		if !ok {
			return nil, nil, fmt.Errorf("%s: unknown language %q", path, lang)
		}
		if err := p.ts.SetLanguage(g); err != nil {
			return nil, nil, fmt.Errorf("%s: %w", path, err)
		}
		p.current = lang
	}
	tree, err := p.ts.Parse(src)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", path, err)
	}
	if err := p.w.run(f, src, tree.RootNode(), p.limits); err != nil {
		return nil, nil, err
	}
	assert.That(len(f.Stripped) == len(f.Lines), "the stripped text keeps every line")
	return f, tree, nil
}

// splitLines splits src at each newline. A final newline gives no extra line.
func splitLines(src []byte) []string {
	trimmed := bytes.TrimSuffix(src, []byte("\n"))
	parts := bytes.Split(trimmed, []byte("\n"))
	out := make([]string, len(parts))
	for i, part := range parts {
		out[i] = string(bytes.TrimSuffix(part, []byte("\r")))
	}
	assert.That(len(out) >= 1, "a file has one or more lines")
	assert.That(len(out) == bytes.Count(trimmed, []byte("\n"))+1, "one entry for each line")
	return out
}
