package mutation

import (
	"fmt"
	"os"
	"path/filepath"

	gitpkg "github.com/prabs/prabs-go/internal/git"
)

// Sandbox represents an isolated copy of a repository, safe to mutate.
type Sandbox struct {
	Original string
	Dir      string
}

// NewSandbox creates a fresh sandbox rooted at a temp directory
// containing a copy of src (excluding .git). Original tree is never modified.
func NewSandbox(src string) (*Sandbox, error) {
	abs, err := filepath.Abs(src)
	if err != nil {
		return nil, err
	}
	tmp, err := os.MkdirTemp("", "prabs-mutation-")
	if err != nil {
		return nil, err
	}
	if err := gitpkg.CopyTree(abs, tmp); err != nil {
		os.RemoveAll(tmp)
		return nil, err
	}
	return &Sandbox{Original: abs, Dir: tmp}, nil
}

// Cleanup removes the sandbox. Refuses to remove the original.
func (s *Sandbox) Cleanup() error {
	if s == nil || s.Dir == "" {
		return nil
	}
	if s.Dir == s.Original {
		return fmt.Errorf("refusing to delete original tree")
	}
	// safety: only remove if under os.TempDir
	if !isUnder(s.Dir, os.TempDir()) {
		return fmt.Errorf("refusing to delete sandbox outside temp dir: %s", s.Dir)
	}
	return os.RemoveAll(s.Dir)
}

func isUnder(path, parent string) bool {
	rel, err := filepath.Rel(parent, path)
	if err != nil {
		return false
	}
	return rel != "" && rel[0] != '.'
}
