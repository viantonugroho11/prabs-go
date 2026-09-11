package config

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/prabs/prabs-go/internal/analyzer"
)

type Config struct {
	Version int
	Rules   map[string]analyzer.RuleConfig
	Exclude []string
	FailOn  analyzer.Severity
}

func Default() *Config {
	return &Config{
		Version: 1,
		Rules: map[string]analyzer.RuleConfig{
			"PRABS-CPLX-001": {Enabled: true, Options: map[string]any{"threshold": 10}},
			"PRABS-FUNC-001": {Enabled: true, Options: map[string]any{"threshold": 100}},
			"PRABS-FUNC-002": {Enabled: true, Options: map[string]any{"threshold": 5}},
			"PRABS-NEST-001": {Enabled: true, Options: map[string]any{"threshold": 4}},
			"PRABS-DUP-001":  {Enabled: true, Options: map[string]any{"min_statements": 5}},
			"PRABS-ERR-001":  {Enabled: true, Options: map[string]any{}},
			"PRABS-ERR-002":  {Enabled: true, Options: map[string]any{}},
			"PRABS-NAME-001": {Enabled: true, Options: map[string]any{}},
			"PRABS-FILE-001": {Enabled: true, Options: map[string]any{"threshold": 500}},
		},
		Exclude: []string{"vendor", "testdata", "generated"},
		FailOn:  analyzer.SeverityHigh,
	}
}

// Load reads a minimal YAML config from path.
// Grammar supported:
//
//	version: <int>
//	rules:
//	  RULE-ID:
//	    enabled: true|false
//	    <key>: <int|string>
//	exclude:
//	  - name
//	severity:
//	  fail_on: high
func Load(path string) (*Config, error) {
	cfg := Default()
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var (
		section     string // "", "rules", "exclude", "severity"
		currentRule string
	)
	for scanner.Scan() {
		raw := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(raw), "#") {
			continue
		}
		line := strings.TrimRight(raw, " \t")
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := leadingSpaces(line)
		trim := strings.TrimSpace(line)

		switch indent {
		case 0:
			switch {
			case strings.HasPrefix(trim, "version:"):
				v := strings.TrimSpace(strings.TrimPrefix(trim, "version:"))
				if n, err := strconv.Atoi(v); err == nil {
					cfg.Version = n
				}
			case trim == "rules:":
				section = "rules"
				currentRule = ""
			case trim == "exclude:":
				section = "exclude"
				cfg.Exclude = nil
			case trim == "severity:":
				section = "severity"
			default:
				section = ""
			}
		case 2:
			if section == "rules" && strings.HasSuffix(trim, ":") {
				currentRule = strings.TrimSuffix(trim, ":")
				if _, ok := cfg.Rules[currentRule]; !ok {
					cfg.Rules[currentRule] = analyzer.RuleConfig{Enabled: true, Options: map[string]any{}}
				}
			} else if section == "exclude" && strings.HasPrefix(trim, "- ") {
				cfg.Exclude = append(cfg.Exclude, strings.TrimSpace(trim[2:]))
			} else if section == "severity" {
				k, v, ok := splitKV(trim)
				if ok && k == "fail_on" {
					if sv, ok := analyzer.ParseSeverity(v); ok {
						cfg.FailOn = sv
					}
				}
			}
		case 4:
			if section == "rules" && currentRule != "" {
				k, v, ok := splitKV(trim)
				if !ok {
					continue
				}
				rc := cfg.Rules[currentRule]
				if rc.Options == nil {
					rc.Options = map[string]any{}
				}
				switch k {
				case "enabled":
					rc.Enabled = v == "true"
				default:
					if n, err := strconv.Atoi(v); err == nil {
						rc.Options[k] = n
					} else {
						rc.Options[k] = v
					}
				}
				cfg.Rules[currentRule] = rc
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func leadingSpaces(s string) int {
	n := 0
	for _, r := range s {
		if r == ' ' {
			n++
		} else {
			break
		}
	}
	return n
}

func splitKV(s string) (string, string, bool) {
	i := strings.IndexByte(s, ':')
	if i < 0 {
		return "", "", false
	}
	return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1:]), true
}
