package rules

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/prabs/prabs-go/internal/analyzer"
)

type Complexity struct{}

func (Complexity) ID() string          { return "PRABS-CPLX-001" }
func (Complexity) Name() string        { return "High Cyclomatic Complexity" }
func (Complexity) Description() string { return "Function has too many branching paths." }
func (Complexity) Severity() analyzer.Severity {
	return analyzer.SeverityHigh
}

func (c Complexity) Analyze(ctx *analyzer.Context) []analyzer.Finding {
	threshold := ctx.Config.Int("threshold", 10)
	var findings []analyzer.Finding
	for _, decl := range ctx.File.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Body == nil {
			continue
		}
		score := 1
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.IfStmt:
				score++
			case *ast.ForStmt, *ast.RangeStmt:
				score++
			case *ast.CaseClause:
				if len(x.List) > 0 {
					score++
				}
			case *ast.CommClause:
				score++
			case *ast.BinaryExpr:
				if x.Op == token.LAND || x.Op == token.LOR {
					score++
				}
			}
			return true
		})
		if score > threshold {
			pos := ctx.FileSet.Position(fd.Pos())
			findings = append(findings, analyzer.Finding{
				Message:     fmt.Sprintf("Function %s has cyclomatic complexity %d (threshold %d).", fd.Name.Name, score, threshold),
				Description: "High cyclomatic complexity makes code harder to test and maintain.",
				Suggestion:  "Extract helpers or split responsibilities.",
				File:        ctx.FilePath,
				StartLine:   pos.Line,
				StartColumn: pos.Column,
				EndLine:     ctx.FileSet.Position(fd.End()).Line,
				Confidence:  0.9,
			})
		}
	}
	return findings
}
