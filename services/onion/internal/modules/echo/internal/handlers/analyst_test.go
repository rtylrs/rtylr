package handlers

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMarshalGroundingRowsKeepsValidJSONWhenTruncated(t *testing.T) {
	rows := []map[string]any{
		{"name": "Alpha", "revenue": 100},
		{"name": "Beta", "revenue": 200},
		{"name": "Gamma", "revenue": 300},
	}

	got, truncated := marshalGroundingRows(rows, 64)
	if !truncated {
		t.Fatalf("expected rows to be truncated, got %s", got)
	}

	var decoded []map[string]any
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("expected valid JSON after truncation, got %q: %v", got, err)
	}
	if len(decoded) == 0 || len(decoded) >= len(rows) {
		t.Fatalf("expected a non-empty partial row set, got %d rows in %s", len(decoded), got)
	}
}

func TestBuildAnalystGroundingPromptStatesEvidenceRules(t *testing.T) {
	got := buildAnalystGroundingPrompt(
		[]map[string]any{{"revenue": 123.45}},
		`{"view":"ai_invoices","aggregations":[{"func":"sum","col":"total","as":"revenue"}]}`,
	)

	mustContain := []string{
		"Authoritative analyst grounding:",
		"Scope: the authenticated user's own tenant only.",
		"Row count returned: 1",
		"Rows JSON:",
		`"revenue":123.45`,
		"using ONLY this grounding",
		"do not invent missing values",
	}
	for _, want := range mustContain {
		if !strings.Contains(got, want) {
			t.Fatalf("grounding prompt missing %q\nprompt:\n%s", want, got)
		}
	}
}
