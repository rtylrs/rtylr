package application

import (
	"context"
	"math"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/ai/classifier"
)

func fieldEntries(t *testing.T, out map[string]any) map[string]map[string]any {
	t.Helper()
	raw, ok := out["fields"].([]map[string]any)
	if !ok {
		t.Fatalf("fields output has type %T", out["fields"])
	}
	byID := map[string]map[string]any{}
	for _, entry := range raw {
		id, _ := entry["id"].(string)
		byID[id] = entry
	}
	return byID
}

func TestDocumentFormMappingMapsWithLexicalFallback(t *testing.T) {
	exec := NewDocumentFormMappingToolExecutor(nil, nil)
	input := map[string]any{
		"document": "Legal name: Sofia Larsson\nPhone number: (503) 555-0142\nInsurance: Yes\n",
		"form":     "Patient intake",
		"fields": []any{
			map[string]any{"id": "name", "label": "Legal name"},
			map[string]any{"id": "phone", "label": "Phone number"},
			map[string]any{"id": "insurance", "label": "Has insurance", "role": "Checkbox"},
			map[string]any{"id": "notes", "label": "Internal notes"},
		},
	}
	out, executed, err := exec.ExecuteCaretTool(context.Background(), DocumentFormMappingToolName, input)
	if err != nil {
		t.Fatalf("ExecuteCaretTool: %v", err)
	}
	if !executed {
		t.Fatal("executor did not claim the tool")
	}
	if out["scorer"] != "lexical" {
		t.Fatalf("scorer = %v", out["scorer"])
	}
	if out["field_count"] != 4 {
		t.Fatalf("field_count = %v", out["field_count"])
	}
	fields := fieldEntries(t, out)
	if got := fields["phone"]["action"]; got != "fill" {
		t.Fatalf("phone action = %v", got)
	}
	if got := fields["phone"]["value"]; got != "(503) 555-0142" {
		t.Fatalf("phone value = %v", got)
	}
	if got := fields["insurance"]["action"]; got != "check" {
		t.Fatalf("insurance action = %v", got)
	}
	if got := fields["notes"]["action"]; got != "skip" {
		t.Fatalf("notes action = %v", got)
	}
	if summary, _ := out["summary"].(string); summary == "" {
		t.Fatal("summary is empty")
	}
}

func TestDocumentFormMappingAcceptsPreExtractedEntities(t *testing.T) {
	exec := NewDocumentFormMappingToolExecutor(nil, nil)
	input := map[string]any{
		"entities": []any{"Vendor: Acme Supply", map[string]any{"label": "Total", "value": "1,240.50"}},
		"fields":   []any{map[string]any{"id": "v", "label": "Vendor"}},
	}
	out, executed, err := exec.ExecuteCaretTool(context.Background(), DocumentFormMappingToolName, input)
	if err != nil || !executed {
		t.Fatalf("executed=%v err=%v", executed, err)
	}
	if out["extracted_entity_count"] != 2 {
		t.Fatalf("extracted_entity_count = %v", out["extracted_entity_count"])
	}
	if fields := fieldEntries(t, out); fields["v"]["action"] != "fill" {
		t.Fatalf("vendor action = %v", fields["v"]["action"])
	}
}

func TestDocumentFormMappingSkipsWithoutEntities(t *testing.T) {
	exec := NewDocumentFormMappingToolExecutor(nil, nil)
	out, executed, err := exec.ExecuteCaretTool(context.Background(), DocumentFormMappingToolName, map[string]any{
		"document": "no label value pairs in this document",
		"fields":   []any{map[string]any{"id": "a", "label": "Anything"}},
	})
	if err != nil || !executed {
		t.Fatalf("executed=%v err=%v", executed, err)
	}
	if out["notice"] == nil {
		t.Fatal("expected a notice when nothing was extracted")
	}
	if fields := fieldEntries(t, out); fields["a"]["action"] != "skip" || fields["a"]["needs_review"] != true {
		t.Fatalf("decision = %+v", fields["a"])
	}
}

func TestDocumentFormMappingRejectsBadInput(t *testing.T) {
	exec := NewDocumentFormMappingToolExecutor(nil, nil)
	cases := []map[string]any{
		{"fields": []any{map[string]any{"label": "Phone"}}},                                         // no document or entities
		{"document": "Phone: 555", "fields": []any{}},                                               // no fields
		{"document": "Phone: 555", "fields": []any{map[string]any{"label": ""}}},                    // empty label
		{"document": "Phone: 555", "fields": []any{map[string]any{"label": "x"}}, "threshold": 2.0}, // bad threshold
		{"document": "Phone: 555", "fields": []any{map[string]any{"label": "x"}}, "entities": []any{"no colon"}},
	}
	for i, input := range cases {
		if _, executed, err := exec.ExecuteCaretTool(context.Background(), DocumentFormMappingToolName, input); err == nil || !executed {
			t.Errorf("case %d: executed=%v err=%v, want an error", i, executed, err)
		}
	}
}

func TestDocumentFormMappingRejectsOversizedInput(t *testing.T) {
	exec := NewDocumentFormMappingToolExecutor(nil, nil)
	big := make([]byte, documentFormMappingMaxDocumentBytes+1)
	for i := range big {
		big[i] = 'a'
	}
	fields := make([]any, documentFormMappingMaxFields+1)
	for i := range fields {
		fields[i] = map[string]any{"label": "Field"}
	}
	cases := []map[string]any{
		{"document": string(big), "fields": []any{map[string]any{"label": "x"}}},
		{"document": "x: y", "fields": fields},
	}
	for i, input := range cases {
		if _, _, err := exec.ExecuteCaretTool(context.Background(), DocumentFormMappingToolName, input); err == nil {
			t.Errorf("case %d: expected an error", i)
		}
	}
}

func TestDocumentFormMappingIgnoresOtherTools(t *testing.T) {
	exec := NewDocumentFormMappingToolExecutor(nil, nil)
	out, executed, err := exec.ExecuteCaretTool(context.Background(), "erp.list_products", nil)
	if out != nil || executed || err != nil {
		t.Fatalf("out=%v executed=%v err=%v, want an unclaimed tool", out, executed, err)
	}
}

func TestDocumentFormMappingSubmitNeedsAuthorization(t *testing.T) {
	exec := NewDocumentFormMappingToolExecutor(nil, nil)
	input := map[string]any{
		"document": "Phone number: (503) 555-0142",
		"fields": []any{
			map[string]any{"id": "submit", "label": "Submit", "role": "Button"},
		},
		"allow_submit": true,
	}
	out, executed, err := exec.ExecuteCaretTool(context.Background(), DocumentFormMappingToolName, input)
	if err != nil || !executed {
		t.Fatalf("executed=%v err=%v", executed, err)
	}
	submit, ok := out["submit"].(map[string]any)
	if !ok {
		t.Fatalf("expected a submit entry, got %v", out["submit"])
	}
	if submit["requires_approval"] != true {
		t.Fatalf("submit = %+v", submit)
	}
}

func TestDocumentFormMappingSubmitDemotedWithoutAuthorization(t *testing.T) {
	exec := NewDocumentFormMappingToolExecutor(nil, nil)
	out, executed, err := exec.ExecuteCaretTool(context.Background(), DocumentFormMappingToolName, map[string]any{
		"document": "Phone number: (503) 555-0142",
		"fields":   []any{map[string]any{"id": "submit", "label": "Submit", "role": "Button"}},
	})
	if err != nil || !executed {
		t.Fatalf("executed=%v err=%v", executed, err)
	}
	if out["submit"] != nil {
		t.Fatalf("submit planned without authorization: %v", out["submit"])
	}
	if entry := fieldEntries(t, out)["submit"]; entry["action"] != "skip" || entry["needs_review"] != true {
		t.Fatalf("decision = %+v, want skip with review", entry)
	}
}

func TestDocumentFormMappingKeepsOneSubmit(t *testing.T) {
	exec := NewDocumentFormMappingToolExecutor(nil, nil)
	out, executed, err := exec.ExecuteCaretTool(context.Background(), DocumentFormMappingToolName, map[string]any{
		"document":     "Phone number: (503) 555-0142",
		"allow_submit": true,
		"fields": []any{
			map[string]any{"id": "a", "label": "Submit", "role": "Button"},
			map[string]any{"id": "b", "label": "Submit Form", "role": "Button"},
		},
	})
	if err != nil || !executed {
		t.Fatalf("executed=%v err=%v", executed, err)
	}
	if out["submit"] == nil {
		t.Fatal("expected one submit click")
	}
	clicks := 0
	for _, entry := range fieldEntries(t, out) {
		if entry["action"] == "click" {
			clicks++
		}
	}
	if clicks != 1 {
		t.Fatalf("planned %d submit clicks, want 1", clicks)
	}
}

func TestDocumentFormMappingAlreadyFilledIsNoop(t *testing.T) {
	exec := NewDocumentFormMappingToolExecutor(nil, nil)
	out, executed, err := exec.ExecuteCaretTool(context.Background(), DocumentFormMappingToolName, map[string]any{
		"document": "Legal name: Sofia Larsson",
		"fields":   []any{map[string]any{"id": "name", "label": "Legal name", "value": "Sofia Larsson"}},
	})
	if err != nil || !executed {
		t.Fatalf("executed=%v err=%v", executed, err)
	}
	entry := fieldEntries(t, out)["name"]
	if entry["action"] != "skip" || entry["already_set"] != true || entry["needs_review"] != nil {
		t.Fatalf("decision = %+v, want already-set skip", entry)
	}
}

func TestDocumentFormMappingThresholdDemotesRealScorer(t *testing.T) {
	scorer, err := classifier.Load("../../../../ai/classifier/models/cua-s1-forms")
	if err != nil {
		t.Fatalf("load real checkpoint: %v", err)
	}
	exec := NewDocumentFormMappingToolExecutor(scorer, map[string]any{"kind": "s1-forms"})
	out, executed, err := exec.ExecuteCaretTool(context.Background(), DocumentFormMappingToolName, map[string]any{
		"document":  "Legal name: Sofia Larsson\nPhone number: (503) 555-0142",
		"threshold": 1.0,
		"fields": []any{
			map[string]any{"id": "name", "label": "Legal name"},
			map[string]any{"id": "phone", "label": "Phone number"},
		},
	})
	if err != nil || !executed {
		t.Fatalf("executed=%v err=%v", executed, err)
	}
	if out["fill_count"] != 0 || out["check_count"] != 0 {
		t.Fatalf("threshold 1.0 should demote every decision: fill=%v check=%v", out["fill_count"], out["check_count"])
	}
	if out["needs_review_count"] != 2 {
		t.Fatalf("needs_review_count = %v, want 2", out["needs_review_count"])
	}
}

func TestDocumentFormMappingEnforcesCaps(t *testing.T) {
	exec := NewDocumentFormMappingToolExecutor(nil, nil)
	tooManyFields := make([]any, documentFormMappingMaxFields+1)
	for i := range tooManyFields {
		tooManyFields[i] = map[string]any{"label": "Field"}
	}
	tooManyEntities := make([]any, documentFormMappingMaxEntities+1)
	for i := range tooManyEntities {
		tooManyEntities[i] = "Field: value"
	}
	budgetFields := make([]any, documentFormMappingMaxFields)
	for i := range budgetFields {
		budgetFields[i] = map[string]any{"label": "Field"}
	}
	budgetEntities := make([]any, documentFormMappingMaxEntities)
	for i := range budgetEntities {
		budgetEntities[i] = "Field: value"
	}
	cases := []map[string]any{
		{"document": "x: y", "fields": tooManyFields},
		{"document": "x: y", "fields": []any{map[string]any{"label": "x"}}, "entities": tooManyEntities},
		{"fields": budgetFields, "entities": budgetEntities},
	}
	for i, input := range cases {
		if _, executed, err := exec.ExecuteCaretTool(context.Background(), DocumentFormMappingToolName, input); err == nil || !executed {
			t.Errorf("case %d: executed=%v err=%v, want an error", i, executed, err)
		}
	}
}

func TestDocumentFormMappingTrimsPreExtractedEntities(t *testing.T) {
	exec := NewDocumentFormMappingToolExecutor(nil, nil)
	label := strings.Repeat("é", 100)
	value := strings.Repeat("v", 300)
	out, executed, err := exec.ExecuteCaretTool(context.Background(), DocumentFormMappingToolName, map[string]any{
		"entities": []any{map[string]any{"label": label, "value": value}},
		"fields":   []any{map[string]any{"id": "a", "label": label}},
	})
	if err != nil || !executed {
		t.Fatalf("executed=%v err=%v", executed, err)
	}
	if out["extracted_entity_count"] != 1 {
		t.Fatalf("extracted_entity_count = %v", out["extracted_entity_count"])
	}
	if entry := fieldEntries(t, out)["a"]; entry["action"] != "fill" {
		t.Fatalf("decision = %+v, want fill from the trimmed entity", entry)
	}
}

func TestDocumentFormMappingSkipsURLNoiseEntities(t *testing.T) {
	exec := NewDocumentFormMappingToolExecutor(nil, nil)
	out, executed, err := exec.ExecuteCaretTool(context.Background(), DocumentFormMappingToolName, map[string]any{
		"entities": []any{"See https://example.com/x"},
		"fields":   []any{map[string]any{"id": "a", "label": "Phone"}},
	})
	if err != nil || !executed {
		t.Fatalf("executed=%v err=%v", executed, err)
	}
	if out["extracted_entity_count"] != 0 {
		t.Fatalf("URL noise became an entity: %v", out["extracted_entity_count"])
	}
	if out["notice"] == nil {
		t.Fatal("expected a notice when every entity was dropped")
	}
}

func TestDocumentFormMappingTrimsFieldLengths(t *testing.T) {
	exec := NewDocumentFormMappingToolExecutor(nil, nil)
	longLabel := strings.Repeat("é", 120)
	out, executed, err := exec.ExecuteCaretTool(context.Background(), DocumentFormMappingToolName, map[string]any{
		"document": "Phone: (503) 555-0142",
		"fields":   []any{map[string]any{"id": "a", "label": longLabel, "value": strings.Repeat("v", 400)}},
	})
	if err != nil || !executed {
		t.Fatalf("executed=%v err=%v", executed, err)
	}
	entry := fieldEntries(t, out)["a"]
	label, _ := entry["label"].(string)
	if len(label) > 80 {
		t.Fatalf("field label not trimmed: %d bytes", len(label))
	}
}

func TestDocumentFormMappingRejectsNonFiniteThreshold(t *testing.T) {
	exec := NewDocumentFormMappingToolExecutor(nil, nil)
	for name, threshold := range map[string]float64{"nan": math.NaN(), "inf": math.Inf(1), "neg_inf": math.Inf(-1)} {
		_, executed, err := exec.ExecuteCaretTool(context.Background(), DocumentFormMappingToolName, map[string]any{
			"document":  "Phone: (503) 555-0142",
			"fields":    []any{map[string]any{"label": "Phone"}},
			"threshold": threshold,
		})
		if err == nil || !executed {
			t.Errorf("%s threshold: executed=%v err=%v, want an error", name, executed, err)
		}
	}
}

func TestDocumentFormMappingKeepsPaddedURLValue(t *testing.T) {
	exec := NewDocumentFormMappingToolExecutor(nil, nil)
	out, executed, err := exec.ExecuteCaretTool(context.Background(), DocumentFormMappingToolName, map[string]any{
		"entities": []any{"  Site :  https://example.com  "},
		"fields":   []any{map[string]any{"id": "site", "label": "Site"}},
	})
	if err != nil || !executed {
		t.Fatalf("executed=%v err=%v", executed, err)
	}
	if out["extracted_entity_count"] != 1 {
		t.Fatalf("padded clean URL was dropped: %v", out["extracted_entity_count"])
	}
	if entry := fieldEntries(t, out)["site"]; entry["action"] != "fill" || entry["value"] != "https://example.com" {
		t.Fatalf("decision = %+v, want fill with the trimmed URL", entry)
	}
}
