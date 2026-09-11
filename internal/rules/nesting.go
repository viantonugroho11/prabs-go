package rules

import (
	"fmt"
	"go/ast"

	"github.com/prabs/prabs-go/internal/analyzer"
)

type DeepNesting struct{}

func (DeepNesting) ID() string          { return "PRABS-NEST-001" }
func (DeepNesting) Name() string        { return "Deep Nesting" }
func (DeepNesting) Description() string { return "Control-flow nesting is too deep." }
func (DeepNesting) Severity() analyzer.Severity {
	return analyzer.SeverityMedium
}

func maxDepth(n ast.Node, depth int) int {
	if n == nil {
		return depth
	}
	best := depth
	ast.Inspect(n, func(x ast.Node) bool {
		return false // manual walk
	})
	// manual traversal via type switch
	switch x := n.(type) {
	case *ast.BlockStmt:
		for _, s := range x.List {
			if d := maxDepth(s, depth); d > best {
				best = d
			}
		}
	case *ast.IfStmt:
		if d := maxDepth(x.Body, depth+1); d > best {
			best = d
		}
		if x.Else != nil {
			if d := maxDepth(x.Else, depth+1); d > best {
				best = d
			}
		}
	case *ast.ForStmt:
		if d := maxDepth(x.Body, depth+1); d > best {
			best = d
		}
	case *ast.RangeStmt:
		if d := maxDepth(x.Body, depth+1); d > best {
			best = d
		}
	case *ast.SwitchStmt:
		if d := maxDepth(x.Body, depth+1); d > best {
			best = d
		}
	case *ast.TypeSwitchStmt:
		if d := maxDepth(x.Body, depth+1); d > best {
			best = d
		}
	case *ast.SelectStmt:
		if d := maxDepth(x.Body, depth+1); d > best {
			best = d
		}
	case *ast.CaseClause:
		for _, s := range x.Body {
			if d := maxDepth(s, depth); d > best {
				best = d
			}
		}
	case *ast.CommClause:
		for _, s := range x.Body {
			if d := maxDepth(s, depth); d > best {
				best = d
			}
		}
	}
	return best
}

func (r DeepNesting) Analyze(ctx *analyzer.Context) []analyzer.Finding {
	threshold := ctx.Config.Int("threshold", 4)
	var findings []analyzer.Finding
	for _, decl := range ctx.File.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Body == nil {
			continue
		}
		depth := maxDepth(fd.Body, 0)
		if depth > threshold {
			pos := ctx.FileSet.Position(fd.Pos())
			findings = append(findings, analyzer.Finding{
				Message:    fmt.Sprintf("Function %s has nesting depth %d (threshold %d).", fd.Name.Name, depth, threshold),
				Suggestion: "Use early returns or extract inner blocks.",
				File:       ctx.FilePath,
				StartLine:  pos.Line,
				Confidence: 0.85,
			})
		}
	}
	return findings
}
