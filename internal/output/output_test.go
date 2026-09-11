package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/prabs/prabs-go/internal/analyzer"
)

func sampleResult() *analyzer.Result {
	return &analyzer.Result{
		Findings: []analyzer.Finding{{
			RuleID: "PRABS-CPLX-001", RuleName: "cx", Severity: analyzer.SeverityHigh,
			Message: "boom", File: "x.go", StartLine: 3, StartColumn: 1,
		}},
		FilesAnalyzed: 1,
		RulesExecuted: 1,
	}
}

func TestJSONOutput(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, FormatJSON, sampleResult()); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["files_analyzed"].(float64) != 1 {
		t.Fatal("files_analyzed missing")
	}
}

func TestSARIFOutput(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, FormatSARIF, sampleResult()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "\"version\": \"2.1.0\"") {
		t.Fatal("expected SARIF version")
	}
}

func TestTextOutput(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, FormatText, sampleResult()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "PRABS-CPLX-001") {
		t.Fatal("rule id missing")
	}
}
