package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/prabs/prabs-go/internal/analyzer"
)

type Format string

const (
	FormatText  Format = "text"
	FormatJSON  Format = "json"
	FormatSARIF Format = "sarif"
)

func Write(w io.Writer, format Format, res *analyzer.Result) error {
	switch format {
	case FormatJSON:
		return writeJSON(w, res)
	case FormatSARIF:
		return writeSARIF(w, res)
	default:
		return writeText(w, res)
	}
}

func writeText(w io.Writer, res *analyzer.Result) error {
	if len(res.Findings) == 0 {
		fmt.Fprintln(w, "No findings.")
	}
	for _, f := range res.Findings {
		fmt.Fprintf(w, "%s:%d:%d\n[%s] %s\n%s\n",
			f.File, f.StartLine, f.StartColumn,
			strings.ToUpper(string(f.Severity)), f.RuleID, f.Message)
		if f.Suggestion != "" {
			fmt.Fprintf(w, "  Suggestion: %s\n", f.Suggestion)
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintf(w, "Summary: %d finding(s), %d file(s), %d rule(s), %s\n",
		len(res.Findings), res.FilesAnalyzed, res.RulesExecuted, res.Duration)
	return nil
}

func writeJSON(w io.Writer, res *analyzer.Result) error {
	payload := map[string]any{
		"findings":       res.Findings,
		"files_analyzed": res.FilesAnalyzed,
		"rules_executed": res.RulesExecuted,
		"duration_ms":    res.Duration.Milliseconds(),
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}

type sarifLog struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []sarifRun `json:"runs"`
}
type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}
type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}
type sarifDriver struct {
	Name  string      `json:"name"`
	Rules []sarifRule `json:"rules"`
}
type sarifRule struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifMessage    `json:"message"`
	Locations []sarifLocation `json:"locations"`
}
type sarifMessage struct {
	Text string `json:"text"`
}
type sarifLocation struct {
	PhysicalLocation sarifPhys `json:"physicalLocation"`
}
type sarifPhys struct {
	ArtifactLocation sarifArtifact `json:"artifactLocation"`
	Region           sarifRegion   `json:"region"`
}
type sarifArtifact struct {
	URI string `json:"uri"`
}
type sarifRegion struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn,omitempty"`
	EndLine     int `json:"endLine,omitempty"`
}

func severityToSarif(s analyzer.Severity) string {
	switch s {
	case analyzer.SeverityCritical, analyzer.SeverityHigh:
		return "error"
	case analyzer.SeverityMedium:
		return "warning"
	default:
		return "note"
	}
}

func writeSARIF(w io.Writer, res *analyzer.Result) error {
	rulesSeen := map[string]string{}
	var results []sarifResult
	for _, f := range res.Findings {
		rulesSeen[f.RuleID] = f.RuleName
		startCol := f.StartColumn
		if startCol == 0 {
			startCol = 1
		}
		results = append(results, sarifResult{
			RuleID:  f.RuleID,
			Level:   severityToSarif(f.Severity),
			Message: sarifMessage{Text: f.Message},
			Locations: []sarifLocation{{
				PhysicalLocation: sarifPhys{
					ArtifactLocation: sarifArtifact{URI: f.File},
					Region:           sarifRegion{StartLine: f.StartLine, StartColumn: startCol, EndLine: f.EndLine},
				},
			}},
		})
	}
	var rules []sarifRule
	for id, name := range rulesSeen {
		rules = append(rules, sarifRule{ID: id, Name: name})
	}
	log := sarifLog{
		Version: "2.1.0",
		Schema:  "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0.json",
		Runs: []sarifRun{{
			Tool:    sarifTool{Driver: sarifDriver{Name: "prabs-go", Rules: rules}},
			Results: results,
		}},
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(log)
}
