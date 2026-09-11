package rules

import (
	"go/ast"

	"github.com/prabs/prabs-go/internal/analyzer"
)

type PoorNaming struct{}

func (PoorNaming) ID() string   { return "PRABS-NAME-001" }
func (PoorNaming) Name() string { return "Poor Naming" }
func (PoorNaming) Description() string {
	return "Identifier uses a low-signal name (foo, bar, tmp, data)."
}
func (PoorNaming) Severity() analyzer.Severity {
	return analyzer.SeverityLow
}

var badNames = map[string]bool{
	"foo": true, "bar": true, "baz": true, "tmp": true, "temp": true, "data": true, "obj": true,
}

func (r PoorNaming) Analyze(ctx *analyzer.Context) []analyzer.Finding {
	var findings []analyzer.Finding
	for _, decl := range ctx.File.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if badNames[fd.Name.Name] {
			pos := ctx.FileSet.Position(fd.Pos())
			findings = append(findings, analyzer.Finding{
				Message:    "Function name '" + fd.Name.Name + "' is low-signal.",
				Suggestion: "Use a domain-meaningful name.",
				File:       ctx.FilePath,
				StartLine:  pos.Line,
				Confidence: 0.9,
			})
		}
	}
	return findings
}
