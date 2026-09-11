package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/prabs/prabs-go/internal/analyzer"
)

type IgnoredError struct{}

func (IgnoredError) ID() string   { return "PRABS-ERR-001" }
func (IgnoredError) Name() string { return "Ignored Error" }
func (IgnoredError) Description() string {
	return "Function returning error is called without checking result."
}
func (IgnoredError) Severity() analyzer.Severity {
	return analyzer.SeverityMedium
}

// Heuristic: flag `foo()` as ExprStmt where the callee identifier suggests error return.
// Also flag `x, _ := f()` where the blank is assigned in the second position (common err pattern).
func (r IgnoredError) Analyze(ctx *analyzer.Context) []analyzer.Finding {
	var findings []analyzer.Finding
	ast.Inspect(ctx.File, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.AssignStmt:
			// look for `_` in RHS assignment: a, _ := f()
			if len(x.Rhs) == 1 {
				if _, ok := x.Rhs[0].(*ast.CallExpr); ok {
					for i, lhs := range x.Lhs {
						if id, ok := lhs.(*ast.Ident); ok && id.Name == "_" && i > 0 {
							pos := ctx.FileSet.Position(x.Pos())
							findings = append(findings, analyzer.Finding{
								Message:    "Error return value is discarded with `_`.",
								Suggestion: "Handle the error or document why it is safe to ignore.",
								File:       ctx.FilePath,
								StartLine:  pos.Line,
								Confidence: 0.7,
							})
						}
					}
				}
			}
		case *ast.ExprStmt:
			// bare call whose name hints at error return
			if call, ok := x.X.(*ast.CallExpr); ok {
				name := calleeName(call)
				if hintsError(name) {
					pos := ctx.FileSet.Position(x.Pos())
					findings = append(findings, analyzer.Finding{
						Message:    "Error return from " + name + " appears to be ignored.",
						Suggestion: "Check and handle the returned error.",
						File:       ctx.FilePath,
						StartLine:  pos.Line,
						Confidence: 0.6,
					})
				}
			}
		}
		return true
	})
	return findings
}

func calleeName(c *ast.CallExpr) string {
	switch fn := c.Fun.(type) {
	case *ast.Ident:
		return fn.Name
	case *ast.SelectorExpr:
		return fn.Sel.Name
	}
	return ""
}

func hintsError(name string) bool {
	if name == "" {
		return false
	}
	low := strings.ToLower(name)
	for _, p := range []string{"read", "write", "open", "close", "encode", "decode", "marshal", "unmarshal", "parse", "exec", "run", "send", "recv", "dial", "commit", "rollback"} {
		if strings.HasPrefix(low, p) {
			return true
		}
	}
	return false
}

// PRABS-ERR-002: fmt.Errorf using %v instead of %w when wrapping an error.
type ErrorWrapping struct{}

func (ErrorWrapping) ID() string          { return "PRABS-ERR-002" }
func (ErrorWrapping) Name() string        { return "Error Wrapping Missing" }
func (ErrorWrapping) Description() string { return "fmt.Errorf uses %v where %w is more appropriate." }
func (ErrorWrapping) Severity() analyzer.Severity {
	return analyzer.SeverityLow
}
func (r ErrorWrapping) Analyze(ctx *analyzer.Context) []analyzer.Finding {
	var findings []analyzer.Finding
	ast.Inspect(ctx.File, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok || pkg.Name != "fmt" || sel.Sel.Name != "Errorf" {
			return true
		}
		if len(call.Args) < 2 {
			return true
		}
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		if strings.Contains(lit.Value, "%v") && !strings.Contains(lit.Value, "%w") {
			// suspect only if any arg is likely err (heuristic by name)
			for _, a := range call.Args[1:] {
				if id, ok := a.(*ast.Ident); ok && strings.Contains(strings.ToLower(id.Name), "err") {
					pos := ctx.FileSet.Position(call.Pos())
					findings = append(findings, analyzer.Finding{
						Message:    "Consider %w instead of %v to wrap the error.",
						Suggestion: "Use fmt.Errorf(\"...: %w\", err) for wrapping.",
						File:       ctx.FilePath,
						StartLine:  pos.Line,
						Confidence: 0.7,
					})
					break
				}
			}
		}
		return true
	})
	return findings
}
