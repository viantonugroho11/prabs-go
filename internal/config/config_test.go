package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadYAMLSubset(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".prabs.yaml")
	body := `version: 1
rules:
  PRABS-CPLX-001:
    enabled: true
    threshold: 15
  PRABS-FUNC-001:
    enabled: false
exclude:
  - vendor
  - generated
severity:
  fail_on: critical
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Rules["PRABS-CPLX-001"].Int("threshold", 0) != 15 {
		t.Fatal("threshold override not applied")
	}
	if cfg.Rules["PRABS-FUNC-001"].Enabled {
		t.Fatal("PRABS-FUNC-001 should be disabled")
	}
	if string(cfg.FailOn) != "critical" {
		t.Fatalf("fail_on = %s", cfg.FailOn)
	}
	if len(cfg.Exclude) != 2 {
		t.Fatalf("exclude count = %d", len(cfg.Exclude))
	}
}

func TestDefault(t *testing.T) {
	c := Default()
	if !c.Rules["PRABS-CPLX-001"].Enabled {
		t.Fatal("default rule should be enabled")
	}
}
