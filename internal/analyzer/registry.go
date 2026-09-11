package analyzer

import "sort"

type Registry struct {
	rules map[string]Rule
}

func NewRegistry() *Registry { return &Registry{rules: map[string]Rule{}} }

func (r *Registry) Register(rule Rule) { r.rules[rule.ID()] = rule }

func (r *Registry) Get(id string) (Rule, bool) {
	rule, ok := r.rules[id]
	return rule, ok
}

func (r *Registry) All() []Rule {
	out := make([]Rule, 0, len(r.rules))
	for _, rule := range r.rules {
		out = append(out, rule)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID() < out[j].ID() })
	return out
}
