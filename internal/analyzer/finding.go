package analyzer

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

var severityRank = map[Severity]int{
	SeverityInfo:     0,
	SeverityLow:      1,
	SeverityMedium:   2,
	SeverityHigh:     3,
	SeverityCritical: 4,
}

func (s Severity) Rank() int { return severityRank[s] }

func ParseSeverity(s string) (Severity, bool) {
	sv := Severity(s)
	if _, ok := severityRank[sv]; ok {
		return sv, true
	}
	return "", false
}

type Finding struct {
	RuleID      string   `json:"rule_id"`
	RuleName    string   `json:"rule_name"`
	Severity    Severity `json:"severity"`
	Confidence  float64  `json:"confidence"`
	Message     string   `json:"message"`
	Description string   `json:"description,omitempty"`
	File        string   `json:"file"`
	StartLine   int      `json:"start_line"`
	StartColumn int      `json:"start_column"`
	EndLine     int      `json:"end_line"`
	EndColumn   int      `json:"end_column"`
	Suggestion  string   `json:"suggestion,omitempty"`
}
