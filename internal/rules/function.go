package rules

import (
	"fmt"
	"go/ast"

	"github.com/prabs/prabs-go/internal/analyzer"
)

type LongFunction struct{}

func (LongFunction) ID() string          { return "PRABS-FUNC-001" }
func (LongFunction) Name() string        { return "Long Function" }
func (LongFunction) Description() string { return "Function body exceeds line threshold." }
func (LongFunction) Severity() analyzer.Severity {
	return analyzer.SeverityMedium
}
func (r LongFunction) Analyze(ctx *analyzer.Context) []analyzer.Finding {
	threshold := ctx.Config.Int("threshold", 100)
	var findings []analyzer.Finding
	for _, decl := range ctx.File.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Body == nil {
			continue
		}
		start := ctx.FileSet.Position(fd.Body.Lbrace).Line
		end := ctx.FileSet.Position(fd.Body.Rbrace).Line
		lines := end - start + 1
		if lines > threshold {
			pos := ctx.FileSet.Position(fd.Pos())
			findings = append(findings, analyzer.Finding{
				Message:    fmt.Sprintf("Function %s is %d lines long (threshold %d).", fd.Name.Name, lines, threshold),
				Suggestion: "Split into smaller focused functions.",
				File:       ctx.FilePath,
				StartLine:  pos.Line,
				EndLine:    end,
				Confidence: 0.95,
			})
		}
	}
	return findings
}

type TooManyParams struct{}

func (TooManyParams) ID() string          { return "PRABS-FUNC-002" }
func (TooManyParams) Name() string        { return "Too Many Parameters" }
func (TooManyParams) Description() string { return "Function has too many parameters." }
func (TooManyParams) Severity() analyzer.Severity {
	return analyzer.SeverityMedium
}
func (r TooManyParams) Analyze(ctx *analyzer.Context) []analyzer.Finding {
	threshold := ctx.Config.Int("threshold", 5)
	var findings []analyzer.Finding
	for _, decl := range ctx.File.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Type.Params == nil {
			continue
		}
		count := 0
		for _, field := range fd.Type.Params.List {
			n := len(field.Names)
			if n == 0 {
				n = 1
			}
			count += n
		}
		if count > threshold {
			pos := ctx.FileSet.Position(fd.Pos())
			findings = append(findings, analyzer.Finding{
				Message:    fmt.Sprintf("Function %s has %d parameters (threshold %d).", fd.Name.Name, count, threshold),
				Suggestion: "Group related parameters into a struct.",
				File:       ctx.FilePath,
				StartLine:  pos.Line,
				Confidence: 0.9,
			})
		}
	}
	return findings
}
