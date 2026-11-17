// Package report writes the findings as SARIF, JSON or text.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/infrashift/semscope/internal/assert"
	"github.com/infrashift/semscope/internal/engine"
)

// Summary counts the findings of each severity.
type Summary struct {
	Files    int `json:"files"`
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Notes    int `json:"notes"`
}

// Summarize counts the findings of r by severity.
func Summarize(r engine.Result) Summary {
	s := Summary{Files: r.Files}
	for _, f := range r.Findings {
		switch f.Severity {
		case "error":
			s.Errors++
		case "warning":
			s.Warnings++
		default:
			s.Notes++
		}
	}
	assert.That(s.Errors+s.Warnings+s.Notes == len(r.Findings), "each finding has one severity")
	assert.That(s.Files >= 0, "the file count comes from the scan")
	return s
}

// WriteText writes one line for each finding, then a summary line:
// "path:line:col: severity: RULE-ID: message". A finding without a location
// starts with "-".
func WriteText(w io.Writer, r engine.Result) error {
	assert.That(w != nil, "w is not nil")
	var b strings.Builder
	for _, f := range r.Findings {
		where := "-"
		if f.Path != "" {
			where = fmt.Sprintf("%s:%d:%d", f.Path, f.Line, f.Col)
		}
		fmt.Fprintf(&b, "%s: %s: %s: %s\n", where, f.Severity, f.RuleID, f.Message)
	}
	s := Summarize(r)
	fmt.Fprintf(&b, "semscope: %d finding(s) in %d file(s): %d error(s), %d warning(s), %d note(s)\n",
		len(r.Findings), s.Files, s.Errors, s.Warnings, s.Notes)
	_, err := io.WriteString(w, b.String())
	assert.That(b.Len() > 0, "the text holds the summary line")
	return err
}

// jsonFinding is the JSON form of one finding.
type jsonFinding struct {
	RuleID   string `json:"ruleId"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Path     string `json:"path,omitempty"`
	Line     int    `json:"line,omitempty"`
	Col      int    `json:"column,omitempty"`
	EndLine  int    `json:"endLine,omitempty"`
	EndCol   int    `json:"endColumn,omitempty"`
}

// jsonStats is the JSON form of the assertion statistics of one language.
type jsonStats struct {
	Functions  int `json:"functions"`
	Assertions int `json:"assertions"`
}

// jsonReport is the JSON document that WriteJSON writes.
type jsonReport struct {
	Tool       string               `json:"tool"`
	Version    string               `json:"version"`
	Summary    Summary              `json:"summary"`
	Assertions map[string]jsonStats `json:"assertions"`
	Findings   []jsonFinding        `json:"findings"`
}

// WriteJSON writes the findings and the summary as one indented JSON document.
func WriteJSON(w io.Writer, r engine.Result, toolVersion string) error {
	assert.That(w != nil, "w is not nil")
	doc := jsonReport{
		Tool: "semscope", Version: toolVersion, Summary: Summarize(r),
		Assertions: map[string]jsonStats{}, Findings: []jsonFinding{},
	}
	for lang, st := range r.Stats {
		doc.Assertions[string(lang)] = jsonStats{Functions: st.Functions, Assertions: st.Assertions}
	}
	for _, f := range r.Findings {
		doc.Findings = append(doc.Findings, jsonFinding{
			RuleID: f.RuleID, Severity: f.Severity, Message: f.Message, Path: f.Path,
			Line: f.Line, Col: f.Col, EndLine: f.EndLine, EndCol: f.EndCol,
		})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	assert.That(len(doc.Findings) == len(r.Findings), "the document holds each finding")
	return enc.Encode(doc)
}
