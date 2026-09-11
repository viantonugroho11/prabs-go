package rules

import (
	"go/parser"
	"go/token"
	"os"
	"testing"

	"github.com/prabs/prabs-go/internal/analyzer"
)

func parseCtx(t *testing.T, path string, opts map[string]any) *analyzer.Context {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	return &analyzer.Context{
		FileSet:  fset,
		File:     f,
		FilePath: path,
		Source:   src,
		Config:   analyzer.RuleConfig{Enabled: true, Options: opts},
	}
}

func TestComplexityRule(t *testing.T) {
	ctx := parseCtx(t, "../../testdata/smells/complexity/bad.go", map[string]any{"threshold": 10})
	fs := Complexity{}.Analyze(ctx)
	if len(fs) == 0 {
		t.Fatal("expected complexity finding on bad.go")
	}
	ctx2 := parseCtx(t, "../../testdata/smells/complexity/good.go", map[string]any{"threshold": 10})
	fs2 := Complexity{}.Analyze(ctx2)
	if len(fs2) != 0 {
		t.Fatalf("expected no findings on good.go, got %d", len(fs2))
	}
}

func TestLongFunctionRule(t *testing.T) {
	tmp := t.TempDir() + "/long.go"
	body := "package p\n\nfunc F() {\n"
	for i := 0; i < 150; i++ {
		body += "\t_ = 1\n"
	}
	body += "}\n"
	if err := os.WriteFile(tmp, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := parseCtx(t, tmp, map[string]any{"threshold": 100})
	fs := LongFunction{}.Analyze(ctx)
	if len(fs) == 0 {
		t.Fatal("expected long function finding")
	}
}

func TestTooManyParamsRule(t *testing.T) {
	tmp := t.TempDir() + "/p.go"
	body := "package p\n\nfunc F(a, b, c, d, e, f, g int) {}\n"
	if err := os.WriteFile(tmp, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := parseCtx(t, tmp, map[string]any{"threshold": 5})
	fs := TooManyParams{}.Analyze(ctx)
	if len(fs) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(fs))
	}
}

func TestDeepNestingRule(t *testing.T) {
	tmp := t.TempDir() + "/n.go"
	body := `package p
func F(x int) int {
	if x > 0 {
		for i := 0; i < 1; i++ {
			if i == 0 {
				switch x {
				case 1:
					if x == 1 { return 1 }
				}
			}
		}
	}
	return 0
}
`
	if err := os.WriteFile(tmp, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := parseCtx(t, tmp, map[string]any{"threshold": 4})
	fs := DeepNesting{}.Analyze(ctx)
	if len(fs) == 0 {
		t.Fatal("expected nesting finding")
	}
}

func TestIgnoredErrorRule(t *testing.T) {
	tmp := t.TempDir() + "/e.go"
	body := `package p
func f() (int, error) { return 0, nil }
func F() { v, _ := f(); _ = v }
`
	if err := os.WriteFile(tmp, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := parseCtx(t, tmp, map[string]any{})
	fs := IgnoredError{}.Analyze(ctx)
	if len(fs) == 0 {
		t.Fatal("expected ignored-error finding")
	}
}

func TestErrorWrappingRule(t *testing.T) {
	tmp := t.TempDir() + "/w.go"
	body := `package p
import "fmt"
func F(err error) error { return fmt.Errorf("boom: %v", err) }
`
	if err := os.WriteFile(tmp, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := parseCtx(t, tmp, map[string]any{})
	fs := ErrorWrapping{}.Analyze(ctx)
	if len(fs) != 1 {
		t.Fatalf("expected 1 wrapping finding, got %d", len(fs))
	}
}

func TestPoorNamingRule(t *testing.T) {
	tmp := t.TempDir() + "/name.go"
	body := "package p\nfunc foo() {}\n"
	if err := os.WriteFile(tmp, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := parseCtx(t, tmp, map[string]any{})
	fs := PoorNaming{}.Analyze(ctx)
	if len(fs) != 1 {
		t.Fatal("expected 1 poor-naming finding")
	}
}

func TestLargeFileRule(t *testing.T) {
	tmp := t.TempDir() + "/big.go"
	body := "package p\n"
	for i := 0; i < 600; i++ {
		body += "var _ = 1\n"
	}
	if err := os.WriteFile(tmp, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := parseCtx(t, tmp, map[string]any{"threshold": 500})
	fs := LargeFile{}.Analyze(ctx)
	if len(fs) != 1 {
		t.Fatal("expected large-file finding")
	}
}

func TestDuplicationRule(t *testing.T) {
	tmp := t.TempDir() + "/dup.go"
	body := `package p
func A() int {
	x := 0
	x++
	x++
	x++
	x++
	x++
	return x
}
func B() int {
	x := 0
	x++
	x++
	x++
	x++
	x++
	return x
}
`
	if err := os.WriteFile(tmp, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := parseCtx(t, tmp, map[string]any{"min_statements": 5})
	fs := Duplication{}.Analyze(ctx)
	if len(fs) == 0 {
		t.Fatal("expected duplication finding")
	}
}
