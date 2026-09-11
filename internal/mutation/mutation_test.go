package mutation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/prabs/prabs-go/internal/analyzer"
	"github.com/prabs/prabs-go/internal/config"
	"github.com/prabs/prabs-go/internal/rules"
)

func TestSandboxIsolation(t *testing.T) {
	orig := t.TempDir()
	src := filepath.Join(orig, "main.go")
	if err := os.WriteFile(src, []byte("package m\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sb, err := NewSandbox(orig)
	if err != nil {
		t.Fatal(err)
	}
	defer sb.Cleanup()
	if sb.Dir == orig {
		t.Fatal("sandbox dir must not equal original")
	}
	// modify sandbox
	extra := filepath.Join(sb.Dir, "junk.go")
	if err := os.WriteFile(extra, []byte("package m\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// original must be unchanged (only main.go present)
	entries, err := os.ReadDir(orig)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("original modified: got %d entries", len(entries))
	}
}

func TestSandboxCleanupSafety(t *testing.T) {
	sb := &Sandbox{Original: "/home/x", Dir: "/etc"}
	if err := sb.Cleanup(); err == nil {
		t.Fatal("expected refusal to clean up /etc")
	}
}

func TestAllMutatorsDetected(t *testing.T) {
	orig := t.TempDir()
	// give original a valid package so copy works
	if err := os.WriteFile(filepath.Join(orig, "seed.go"), []byte("package seed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	reg := analyzer.NewRegistry()
	rules.RegisterAll(reg)
	for _, m := range All() {
		t.Run(m.ID(), func(t *testing.T) {
			sb, err := NewSandbox(orig)
			if err != nil {
				t.Fatal(err)
			}
			defer sb.Cleanup()
			mctx := &MutationContext{Dir: sb.Dir}
			if _, err := m.Mutate(mctx); err != nil {
				t.Fatal(err)
			}
			a := analyzer.New(reg)
			res, err := a.Run(analyzer.Options{
				Roots:     []string{sb.Dir},
				Rules:     cfg.Rules,
				OnlyRules: map[string]bool{m.TargetRule(): true},
			})
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, f := range res.Findings {
				if f.RuleID == m.TargetRule() {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("mutator %s did not trigger %s", m.ID(), m.TargetRule())
			}
		})
	}
}
