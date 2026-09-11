package analyzer

type Rule interface {
	ID() string
	Name() string
	Description() string
	Severity() Severity
	Analyze(ctx *Context) []Finding
}
