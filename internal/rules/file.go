package rules

import (
	"bytes"
	"fmt"

	"github.com/prabs/prabs-go/internal/analyzer"
)

type LargeFile struct{}

func (LargeFile) ID() string          { return "PRABS-FILE-001" }
func (LargeFile) Name() string        { return "Large File" }
func (LargeFile) Description() string { return "File exceeds line threshold." }
func (LargeFile) Severity() analyzer.Severity {
	return analyzer.SeverityLow
}
func (r LargeFile) Analyze(ctx *analyzer.Context) []analyzer.Finding {
	threshold := ctx.Config.Int("threshold", 500)
	lines := bytes.Count(ctx.Source, []byte("\n")) + 1
	if lines > threshold {
		return []analyzer.Finding{{
			Message:    fmt.Sprintf("File has %d lines (threshold %d).", lines, threshold),
			Suggestion: "Split into smaller focused files.",
			File:       ctx.FilePath,
			StartLine:  1,
			EndLine:    lines,
			Confidence: 1.0,
		}}
	}
	return nil
}
