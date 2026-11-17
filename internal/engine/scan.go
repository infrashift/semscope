package engine

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/infrashift/semscope/internal/assert"
	"github.com/infrashift/semscope/internal/parser"
)

// source is one file to scan.
type source struct {
	abs  string
	rel  string
	lang parser.Language
}

// fileResult is the outcome of one file in the pool.
type fileResult struct {
	file     *parser.File
	findings []Finding
	err      error
}

// skipDirs are directory names that a scan does not enter: build output,
// dependencies and test fixtures. A hidden directory is also skipped.
func skipDir(name string) bool {
	assert.That(name != "", "a directory has a name")
	assert.That(!strings.ContainsRune(name, filepath.Separator), "a name has no separator")
	switch name {
	case "vendor", "testdata", "node_modules", "zig-cache", "zig-out", "target", "__pycache__":
		return true
	}
	return strings.HasPrefix(name, ".") && name != "." && name != ".."
}

// Scan checks the files under paths. root is the base of the relative paths in the findings.
func (e *Engine) Scan(root string, paths []string) (Result, error) {
	assert.That(len(e.workers) > 0, "the engine is open")
	sources, err := e.collect(root, paths)
	if err != nil {
		return Result{}, err
	}
	results := make([]fileResult, len(sources))
	e.runPool(sources, results)
	files := make([]*parser.File, 0, len(results))
	var findings []Finding
	for i, r := range results {
		if r.err != nil {
			return Result{}, r.err
		}
		files = append(files, r.file)
		findings = append(findings, r.findings...)
		assert.That(r.file.Path == sources[i].rel, "each result belongs to its source")
	}
	global, stats, err := e.globalChecks(files)
	if err != nil {
		return Result{}, err
	}
	findings = append(findings, global...)
	if len(findings) > e.limits.MaxFindings {
		return Result{}, fmt.Errorf("engine: %d findings (max %d)", len(findings), e.limits.MaxFindings)
	}
	sortFindings(findings)
	return Result{Findings: findings, Files: len(sources), Stats: stats}, nil
}

// collector gathers the source files of one scan.
type collector struct {
	e    *Engine
	root string
	seen map[string]bool
	out  []source
}

// collect returns the source files under paths, sorted and without duplicates.
// It keeps only the languages of the active rules.
func (e *Engine) collect(root string, paths []string) ([]source, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("engine: %w", err)
	}
	c := &collector{e: e, root: absRoot, seen: map[string]bool{}}
	for _, given := range paths {
		if err := c.addPath(given); err != nil {
			return nil, err
		}
	}
	sort.Slice(c.out, func(i, j int) bool { return c.out[i].rel < c.out[j].rel })
	assert.That(len(c.out) == len(c.seen), "one entry for each file")
	assert.That(len(c.out) <= e.limits.MaxFiles, "collect enforces MaxFiles")
	return c.out, nil
}

// addPath adds one file, or each file under one directory.
func (c *collector) addPath(given string) error {
	p, err := filepath.Abs(given)
	if err != nil {
		return fmt.Errorf("engine: %w", err)
	}
	info, err := os.Stat(p)
	if err != nil {
		return fmt.Errorf("engine: %w", err)
	}
	if !info.IsDir() {
		return c.add(filepath.Clean(p))
	}
	err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case d.IsDir() && path != p && skipDir(d.Name()):
			return filepath.SkipDir
		case d.IsDir():
			return nil
		}
		return c.add(path)
	})
	if err != nil {
		return fmt.Errorf("engine: %w", err)
	}
	assert.That(filepath.IsAbs(p), "the walk starts at an absolute path")
	assert.That(len(c.out) <= c.e.limits.MaxFiles, "add enforces MaxFiles")
	return nil
}

// add adds the file p when its language has an active rule.
func (c *collector) add(p string) error {
	assert.That(filepath.IsAbs(p), "p is absolute")
	lang, ok := parser.Detect(p)
	if !ok || !c.e.langs[lang] || c.seen[p] {
		return nil
	}
	if len(c.out) >= c.e.limits.MaxFiles {
		return fmt.Errorf("engine: more than %d files", c.e.limits.MaxFiles)
	}
	c.seen[p] = true
	c.out = append(c.out, source{abs: p, rel: relPath(c.root, p), lang: lang})
	assert.That(len(c.out) == len(c.seen), "one entry for each file")
	return nil
}

// relPath returns p relative to root with forward slashes, or p when p is outside root.
func relPath(root, p string) string {
	assert.That(filepath.IsAbs(root) && filepath.IsAbs(p), "collect makes the paths absolute")
	rel, err := filepath.Rel(root, p)
	if err != nil || strings.HasPrefix(rel, "..") {
		rel = p
	}
	out := filepath.ToSlash(filepath.Clean(rel))
	assert.That(out != "", "a relative path is not empty")
	return out
}

// runPool processes each source in the worker pool. results[i] gets the outcome of sources[i].
func (e *Engine) runPool(sources []source, results []fileResult) {
	assert.That(len(sources) == len(results), "one result for each source")
	jobs := make(chan int, len(sources))
	for i := range sources {
		jobs <- i
	}
	close(jobs)
	var wg sync.WaitGroup
	for w := range e.workers {
		wg.Add(1)
		go func(wk worker) {
			defer wg.Done()
			for i := range jobs {
				results[i] = e.processFile(wk, sources[i])
			}
		}(e.workers[w])
	}
	wg.Wait()
	assert.That(len(jobs) == 0, "the pool takes every job")
}

// processFile reads, parses and checks one file.
func (e *Engine) processFile(wk worker, s source) fileResult {
	assert.That(wk.parser != nil && wk.cursor != nil, "the worker is open")
	info, err := os.Stat(s.abs)
	if err != nil {
		return fileResult{err: fmt.Errorf("engine: %w", err)}
	}
	if info.Size() > int64(e.limits.Parser.MaxFileBytes) {
		return fileResult{err: fmt.Errorf("engine: %s: %d bytes (max %d)", s.rel, info.Size(), e.limits.Parser.MaxFileBytes)}
	}
	src, err := os.ReadFile(s.abs)
	if err != nil {
		return fileResult{err: fmt.Errorf("engine: %w", err)}
	}
	f, tree, err := wk.parser.Parse(s.rel, s.lang, src)
	if err != nil {
		return fileResult{err: fmt.Errorf("engine: %w", err)}
	}
	if f.Exempt {
		return fileResult{file: f}
	}
	defer tree.Close()
	findings := e.checkFile(f, src, tree, wk.cursor)
	f.Stripped = nil
	assert.That(f.Path == s.rel, "the file keeps its relative path")
	return fileResult{file: f, findings: findings}
}

// sortFindings sorts by path, line, column, rule and message.
func sortFindings(fs []Finding) {
	sort.Slice(fs, func(i, j int) bool {
		a, b := fs[i], fs[j]
		switch {
		case a.Path != b.Path:
			return a.Path < b.Path
		case a.Line != b.Line:
			return a.Line < b.Line
		case a.Col != b.Col:
			return a.Col < b.Col
		case a.RuleID != b.RuleID:
			return a.RuleID < b.RuleID
		}
		return a.Message < b.Message
	})
	assert.That(len(fs) == 0 || fs[0].Path <= fs[len(fs)-1].Path, "the paths ascend")
}
