package report

import (
	"io"

	sarif "github.com/owenrumney/go-sarif/v3/pkg/report/v22/sarif"

	"github.com/infrashift/semscope/internal/assert"
)

// writeV22 maps run to SARIF 2.2, validates the report and writes it.
func writeV22(w io.Writer, run sarifRun) error {
	assert.That(w != nil, "w is not nil")
	rep := sarif.NewReport()
	r := sarif.NewRunWithInformationURI(toolName, informationURI)
	r.Tool.Driver.Version = &run.toolVersion
	for _, rule := range run.rules {
		d := r.AddRule(rule.id)
		d.ShortDescription = sarif.NewMultiformatMessageString().WithText(rule.short)
		d.FullDescription = sarif.NewMultiformatMessageString().WithText(rule.full)
		d.DefaultConfiguration = sarif.NewReportingConfiguration().WithLevel(rule.level)
	}
	for _, res := range run.results {
		r.AddResult(v22Result(run, res))
	}
	rep.AddRun(r)
	rep.Guid = sarif.Guid(contentGUID(run))
	if err := rep.Validate(); err != nil {
		return err
	}
	assert.That(len(r.Results) == len(run.results), "the run holds each result")
	return rep.PrettyWrite(w)
}

// v22Result maps one result to SARIF 2.2.
func v22Result(run sarifRun, res sarifResult) *sarif.Result {
	f := res.finding
	assert.That(f.RuleID == run.rules[res.ruleIndex].id, "the rule index points at the rule of the finding")
	out := sarif.NewRuleResult(f.RuleID).
		WithRuleIndex(res.ruleIndex).
		WithLevel(run.rules[res.ruleIndex].level).
		WithMessage(sarif.NewTextMessage(f.Message))
	out.Locations = []*sarif.Location{}
	if f.Path != "" {
		sl, sc, el, ec := region(f)
		loc := sarif.NewArtifactLocation().WithURI(f.Path).WithURIBaseID(srcRoot)
		reg := sarif.NewRegion().WithStartLine(sl).WithStartColumn(sc).WithEndLine(el).WithEndColumn(ec)
		out.Locations = append(out.Locations, sarif.NewLocationWithPhysicalLocation(
			sarif.NewPhysicalLocation().WithArtifactLocation(loc).WithRegion(reg)))
	}
	assert.That(res.ruleIndex < len(run.rules), "the rule index is valid")
	return out
}
