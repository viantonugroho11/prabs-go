package rules

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"go/ast"
	"go/printer"

	"github.com/prabs/prabs-go/internal/analyzer"
)

type Duplication struct{}

func (Duplication) ID() string          { return "PRABS-DUP-001" }
func (Duplication) Name() string        { return "Duplicate Code" }
func (Duplication) Description() string { return "Two function bodies are structurally identical." }
func (Duplication) Severity() analyzer.Severity {
	return analyzer.SeverityMedium
}

// Conservative: hash the printed body of each function ≥ minStmts statements; report collisions.
func (r Duplication) Analyze(ctx *analyzer.Context) []analyzer.Finding {
	minStmts := ctx.Config.Int("min_statements", 5)
	seen := map[string]*ast.FuncDecl{}
	var findings []analyzer.Finding
	for _, decl := range ctx.File.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Body == nil || len(fd.Body.List) < minStmts {
			continue
		}
		var buf bytes.Buffer
		if err := printer.Fprint(&buf, ctx.FileSet, fd.Body); err != nil {
			continue
		}
		h := sha1.Sum(buf.Bytes())
		key := fmt.Sprintf("%x", h)
		if prior, ok := seen[key]; ok {
			pos := ctx.FileSet.Position(fd.Pos())
			priorPos := ctx.FileSet.Position(prior.Pos())
			findings = append(findings, analyzer.Finding{
				Message:    fmt.Sprintf("Function %s is a duplicate of %s (line %d).", fd.Name.Name, prior.Name.Name, priorPos.Line),
				Suggestion: "Extract shared logic into a common helper.",
				File:       ctx.FilePath,
				StartLine:  pos.Line,
				Confidence: 0.8,
			})
		} else {
			seen[key] = fd
		}
	}
	return findings
}
