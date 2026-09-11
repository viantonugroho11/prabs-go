package mutation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// writeMutation writes a fresh .go file into ctx.Dir under pkg "prabs_mutation".
func writeMutation(ctx *MutationContext, name, body string) (string, error) {
	dir := filepath.Join(ctx.Dir, "prabs_mutation")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, name)
	content := "package prabs_mutation\n\n" + body
	return path, os.WriteFile(path, []byte(content), 0o644)
}

// ---- Complexity ----
type ComplexityMutator struct{}

func (ComplexityMutator) ID() string                      { return "complexity-injection" }
func (ComplexityMutator) TargetRule() string              { return "PRABS-CPLX-001" }
func (ComplexityMutator) CanMutate(*MutationContext) bool { return true }
func (m ComplexityMutator) Mutate(ctx *MutationContext) (MutationResult, error) {
	var b strings.Builder
	b.WriteString("func InjectedComplexity(a, b, c, d int) int {\n\tr := 0\n")
	for i := 0; i < 14; i++ {
		fmt.Fprintf(&b, "\tif a > %d { r++ }\n", i)
	}
	b.WriteString("\treturn r\n}\n")
	path, err := writeMutation(ctx, "complexity.go", b.String())
	return MutationResult{
		MutatorID: m.ID(), TargetRule: m.TargetRule(), File: path,
		Description: "Injects function with >10 branches.", ExpectedFinding: true,
	}, err
}

// ---- Long Function ----
type LongFunctionMutator struct{}

func (LongFunctionMutator) ID() string                      { return "long-function-injection" }
func (LongFunctionMutator) TargetRule() string              { return "PRABS-FUNC-001" }
func (LongFunctionMutator) CanMutate(*MutationContext) bool { return true }
func (m LongFunctionMutator) Mutate(ctx *MutationContext) (MutationResult, error) {
	var b strings.Builder
	b.WriteString("func InjectedLong() int {\n\tx := 0\n")
	for i := 0; i < 120; i++ {
		fmt.Fprintf(&b, "\tx += %d\n", i)
	}
	b.WriteString("\treturn x\n}\n")
	path, err := writeMutation(ctx, "long.go", b.String())
	return MutationResult{
		MutatorID: m.ID(), TargetRule: m.TargetRule(), File: path,
		Description: "Injects a function with >100 lines.", ExpectedFinding: true,
	}, err
}

// ---- Params ----
type ParamsMutator struct{}

func (ParamsMutator) ID() string                      { return "parameter-injection" }
func (ParamsMutator) TargetRule() string              { return "PRABS-FUNC-002" }
func (ParamsMutator) CanMutate(*MutationContext) bool { return true }
func (m ParamsMutator) Mutate(ctx *MutationContext) (MutationResult, error) {
	body := "func InjectedManyParams(a, b, c, d, e, f, g string) string { return a+b+c+d+e+f+g }\n"
	path, err := writeMutation(ctx, "params.go", body)
	return MutationResult{
		MutatorID: m.ID(), TargetRule: m.TargetRule(), File: path,
		Description: "Injects a function with 7 parameters.", ExpectedFinding: true,
	}, err
}

// ---- Nesting ----
type NestingMutator struct{}

func (NestingMutator) ID() string                      { return "nesting-injection" }
func (NestingMutator) TargetRule() string              { return "PRABS-NEST-001" }
func (NestingMutator) CanMutate(*MutationContext) bool { return true }
func (m NestingMutator) Mutate(ctx *MutationContext) (MutationResult, error) {
	body := `func InjectedDeepNesting(x int) int {
	if x > 0 {
		for i := 0; i < 3; i++ {
			if i > 0 {
				switch x {
				case 1:
					if x == 1 {
						return 1
					}
				}
			}
		}
	}
	return 0
}
`
	path, err := writeMutation(ctx, "nesting.go", body)
	return MutationResult{
		MutatorID: m.ID(), TargetRule: m.TargetRule(), File: path,
		Description: "Injects nesting depth >4.", ExpectedFinding: true,
	}, err
}

// ---- Ignored Error ----
type IgnoredErrorMutator struct{}

func (IgnoredErrorMutator) ID() string                      { return "ignored-error-injection" }
func (IgnoredErrorMutator) TargetRule() string              { return "PRABS-ERR-001" }
func (IgnoredErrorMutator) CanMutate(*MutationContext) bool { return true }
func (m IgnoredErrorMutator) Mutate(ctx *MutationContext) (MutationResult, error) {
	body := `func doRead() (string, error) { return "", nil }

func InjectedIgnored() {
	v, _ := doRead()
	_ = v
}
`
	path, err := writeMutation(ctx, "ignored.go", body)
	return MutationResult{
		MutatorID: m.ID(), TargetRule: m.TargetRule(), File: path,
		Description: "Injects call discarding an error with _.", ExpectedFinding: true,
	}, err
}

// All returns all built-in mutators.
func All() []Mutator {
	return []Mutator{
		ComplexityMutator{},
		LongFunctionMutator{},
		ParamsMutator{},
		NestingMutator{},
		IgnoredErrorMutator{},
	}
}
