package analyzer

import (
	"go/ast"
	"go/token"
)

// Context passed to each rule. Holds parsed AST for one file plus shared config.
type Context struct {
	FileSet   *token.FileSet
	File      *ast.File
	FilePath  string
	Source    []byte
	Config    RuleConfig
	Generated bool
}

// RuleConfig is a bag of options for a specific rule.
type RuleConfig struct {
	Enabled bool
	Options map[string]any
}

func (rc RuleConfig) Int(key string, def int) int {
	if v, ok := rc.Options[key]; ok {
		switch x := v.(type) {
		case int:
			return x
		case int64:
			return int(x)
		case float64:
			return int(x)
		}
	}
	return def
}
