package ruleset

import (
	"fmt"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"

	"github.com/infrashift/semscope/internal/assert"
)

// sourceFile is one CUE source: a built-in ruleset or a file.
type sourceFile struct {
	name string
	data []byte
}

// unifySources unifies the #Ruleset definition with each source of names, in sequence.
// An empty names list loads DefaultRuleset.
func unifySources(ctx *cue.Context, names []string) (cue.Value, []string, error) {
	assert.That(ctx != nil, "ctx is not nil")
	schema := ctx.CompileString(schemaSource, cue.Filename("schema.cue"))
	if err := schema.Err(); err != nil {
		return cue.Value{}, nil, cueError("schema.cue", err)
	}
	v := schema.LookupPath(cue.ParsePath("#Ruleset"))
	if len(names) == 0 {
		names = []string{DefaultRuleset}
	}
	loaded := make([]string, 0, len(names))
	for _, name := range names {
		files, err := resolveSource(name)
		if err != nil {
			return cue.Value{}, nil, err
		}
		if len(loaded)+len(files) > MaxSources {
			return cue.Value{}, nil, fmt.Errorf("ruleset: more than %d source files", MaxSources)
		}
		for _, f := range files {
			src := ctx.CompileBytes(f.data, cue.Filename(f.name))
			if err := src.Err(); err != nil {
				return cue.Value{}, nil, cueError(f.name, err)
			}
			v = v.Unify(src)
			loaded = append(loaded, f.name)
		}
	}
	if err := v.Err(); err != nil {
		return cue.Value{}, nil, cueError("unify", err)
	}
	assert.That(len(loaded) > 0, "one or more sources")
	assert.That(len(loaded) <= MaxSources, "the source count is within MaxSources")
	return v, loaded, nil
}

// resolveSource returns the CUE sources for one --ruleset value: a built-in
// name, a .cue file, or a directory. A directory gives its .cue files in name
// sequence. A built-in name has priority over a path with the same name.
func resolveSource(name string) ([]sourceFile, error) {
	if data, ok := Builtin(name); ok {
		assert.That(len(data) > 0, "a built-in source is not empty")
		return []sourceFile{{name: "builtin:" + name, data: data}}, nil
	}
	info, err := os.Stat(name)
	if err != nil {
		return nil, fmt.Errorf("ruleset: %q is not a built-in ruleset, a file or a directory: %w", name, err)
	}
	if !info.IsDir() {
		f, err := readSource(name)
		if err != nil {
			return nil, err
		}
		return []sourceFile{f}, nil
	}
	entries, err := os.ReadDir(name)
	if err != nil {
		return nil, fmt.Errorf("ruleset: read directory %q: %w", name, err)
	}
	out := make([]sourceFile, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".cue" {
			continue
		}
		if len(out) >= MaxSources {
			return nil, fmt.Errorf("ruleset: directory %q has more than %d .cue files", name, MaxSources)
		}
		f, err := readSource(filepath.Join(name, e.Name()))
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("ruleset: directory %q has no .cue files", name)
	}
	assert.That(len(out) <= MaxSources, "the file count is within MaxSources")
	return out, nil
}

// readSource reads one .cue file of MaxSourceBytes or fewer.
func readSource(path string) (sourceFile, error) {
	info, err := os.Stat(path)
	if err != nil {
		return sourceFile{}, fmt.Errorf("ruleset: %w", err)
	}
	assert.That(!info.IsDir(), "resolveSource passes files only")
	if info.Size() > MaxSourceBytes {
		return sourceFile{}, fmt.Errorf("ruleset: %q has %d bytes (max %d)", path, info.Size(), MaxSourceBytes)
	}
	data, err := os.ReadFile(path) //nolint:gosec // the user names the ruleset file on the command line
	if err != nil {
		return sourceFile{}, fmt.Errorf("ruleset: %w", err)
	}
	assert.That(path != "", "path is not empty")
	return sourceFile{name: path, data: data}, nil
}
