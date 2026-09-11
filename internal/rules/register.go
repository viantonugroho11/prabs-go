package rules

import "github.com/prabs/prabs-go/internal/analyzer"

// All returns every built-in rule.
func All() []analyzer.Rule {
	return []analyzer.Rule{
		Complexity{},
		LongFunction{},
		TooManyParams{},
		DeepNesting{},
		Duplication{},
		IgnoredError{},
		ErrorWrapping{},
		PoorNaming{},
		LargeFile{},
	}
}

// RegisterAll registers built-in rules into r.
func RegisterAll(r *analyzer.Registry) {
	for _, rule := range All() {
		r.Register(rule)
	}
}
