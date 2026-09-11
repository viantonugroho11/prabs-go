package mutation

import "path/filepath"

// MutationContext gives a mutator a writable sandbox rooted at Dir.
type MutationContext struct {
	Dir string // sandbox root — safe to modify
}

// Path returns an absolute path inside the sandbox.
func (m *MutationContext) Path(rel string) string {
	return filepath.Join(m.Dir, rel)
}

type MutationResult struct {
	MutatorID       string
	TargetRule      string
	File            string
	Description     string
	ExpectedFinding bool
}

type Mutator interface {
	ID() string
	TargetRule() string
	CanMutate(ctx *MutationContext) bool
	Mutate(ctx *MutationContext) (MutationResult, error)
}
