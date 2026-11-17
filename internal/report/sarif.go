package report

import (
	"crypto/sha1" //nolint:gosec // UUID version 5 uses SHA-1 by definition (RFC 9562); no security use
	"fmt"
	"io"
	"strings"

	"github.com/infrashift/semscope/internal/assert"
	"github.com/infrashift/semscope/internal/engine"
	"github.com/infrashift/semscope/internal/ruleset"
	"github.com/infrashift/semscope/internal/version"
)

// The fixed fields of the SARIF driver.
const (
	toolName       = "semscope"
	informationURI = "https://github.com/infrashift/semscope"
	srcRoot        = "%SRCROOT%"
)

// sarifRule is one rule of the driver, before the mapping to a SARIF version.
type sarifRule struct {
	id    string
	short string
	full  string
	level string
}

// sarifResult is one result, before the mapping to a SARIF version.
type sarifResult struct {
	ruleIndex int
	finding   engine.Finding
}

// sarifRun is one run, before the mapping to a SARIF version.
type sarifRun struct {
	toolVersion string
	rules       []sarifRule
	results     []sarifResult
}

// WriteSARIF writes res as a SARIF report of the version sarifVersion. The
// driver lists the active rules. The report passes schema validation before
// the writer writes it.
func WriteSARIF(w io.Writer, sarifVersion string, res engine.Result, active []ruleset.Rule, toolVersion string) error {
	assert.That(w != nil, "w is not nil")
	run, err := buildRun(res, active, toolVersion)
	if err != nil {
		return err
	}
	assert.That(len(run.results) == len(res.Findings), "one result for each finding")
	switch sarifVersion {
	case version.SARIF210:
		return writeV210(w, run)
	case version.SARIF22:
		return writeV22(w, run)
	}
	return fmt.Errorf("report: unsupported SARIF version %q", sarifVersion)
}

// buildRun maps the rules and findings to a run. Each finding needs an active rule.
func buildRun(res engine.Result, active []ruleset.Rule, toolVersion string) (sarifRun, error) {
	run := sarifRun{toolVersion: toolVersion}
	index := make(map[string]int, len(active))
	for _, r := range active {
		if !r.Enabled {
			continue
		}
		index[r.ID] = len(run.rules)
		run.rules = append(run.rules, sarifRule{id: r.ID, short: r.Title, full: r.Message, level: r.Severity})
	}
	for _, f := range res.Findings {
		i, ok := index[f.RuleID]
		if !ok {
			return sarifRun{}, fmt.Errorf("report: finding of rule %s, which is not active", f.RuleID)
		}
		run.results = append(run.results, sarifResult{ruleIndex: i, finding: f})
	}
	assert.That(len(run.rules) == len(index), "one rule for each ID")
	assert.That(len(run.results) == len(res.Findings), "one result for each finding")
	return run, nil
}

// urlNamespace is the RFC 9562 namespace ID for URLs, the base of contentGUID.
const urlNamespace = "\x6b\xa7\xb8\x11\x9d\xad\x11\xd1\x80\xb4\x00\xc0\x4f\xd4\x30\xc8"

// contentGUID returns a version 5 UUID of the run content. The same scan
// gives the same GUID, so the report is the same on each run.
func contentGUID(run sarifRun) string {
	var name strings.Builder
	name.WriteString(informationURI + "#" + run.toolVersion)
	for _, r := range run.results {
		f := r.finding
		fmt.Fprintf(&name, "\n%s|%s|%d|%d|%s", f.RuleID, f.Path, f.Line, f.Col, f.Message)
	}
	sum := sha1.Sum([]byte(urlNamespace + name.String())) //nolint:gosec // RFC 9562 version 5; no security use
	sum[6] = (sum[6] & 0x0f) | 0x50
	sum[8] = (sum[8] & 0x3f) | 0x80
	out := fmt.Sprintf("%x-%x-%x-%x-%x", sum[0:4], sum[4:6], sum[6:8], sum[8:10], sum[10:16])
	assert.That(len(out) == 36, "a UUID has 36 characters")
	assert.That(out[14] == '5', "the UUID has version 5")
	return out
}

// region returns the 1-based region of f. An end before the start becomes the start.
func region(f engine.Finding) (startLine, startCol, endLine, endCol int) {
	assert.That(f.Path != "", "only a located finding has a region")
	startLine, startCol = max(f.Line, 1), max(f.Col, 1)
	endLine, endCol = max(f.EndLine, startLine), max(f.EndCol, 1)
	if endLine == startLine && endCol < startCol {
		endCol = startCol
	}
	assert.That(endLine >= startLine, "the region ends after it starts")
	return startLine, startCol, endLine, endCol
}
