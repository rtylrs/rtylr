package handlers

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
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

func TestBuildSpecGenPromptRoutesDailyRevenueToPaidPOSOrders(t *testing.T) {
	got := buildSpecGenPrompt(
		"ai_pos_orders(total, payment_status, paid_at)\nai_invoices(total, issue_date, paid_date)\n",
		"How much revenue today?",
		"2026-06-18",
	)

	mustContain := []string{
		"Tomorrow is 2026-06-19.",
		"prefer paid ai_pos_orders total",
		`"view":"ai_pos_orders"`,
		`"col":"payment_status","op":"=","value":"paid"`,
		`"col":"paid_at","op":">=","value":"2026-06-18"`,
		`"col":"paid_at","op":"<","value":"2026-06-19"`,
		`"value_col":"low_stock_threshold"`,
	}
	for _, want := range mustContain {
		if !strings.Contains(got, want) {
			t.Fatalf("spec prompt missing %q\nprompt:\n%s", want, got)
		}
	}
}

func TestAnalystGroundingUsesBusinessSnapshotForOwnerOverview(t *testing.T) {
	service := &snapshotAnalystService{}
	handler := &GenerateTextStreamHandler{AnalystService: service}

	got, ok := handler.analystGrounding(context.Background(), "How are we doing today?", 11, 22)
	if !ok {
		t.Fatal("expected business snapshot grounding")
	}
	if got.ToolName != "business.snapshot" {
		t.Fatalf("expected business.snapshot tool, got %#v", got)
	}
	for _, want := range []string{
		"Authoritative business snapshot:",
		"business.snapshot tool",
		"today_revenue",
		"low_stock_products",
		"using ONLY this snapshot",
	} {
		if !strings.Contains(got.Prompt, want) {
			t.Fatalf("snapshot prompt missing %q\nprompt:\n%s", want, got.Prompt)
		}
	}
	if len(service.calls) != len(buildBusinessSnapshotQueries(timeNowUTCForTest())) {
		t.Fatalf("expected all snapshot queries to run, got %d calls", len(service.calls))
	}
	if !strings.Contains(got.SpecJSON, `"kind":"business_snapshot"`) {
		t.Fatalf("expected snapshot spec bundle, got %s", got.SpecJSON)
	}
	if !strings.Contains(got.SpecJSON, `"value_col":"low_stock_threshold"`) {
		t.Fatalf("expected low-stock column comparison in spec bundle, got %s", got.SpecJSON)
	}
}

func TestBusinessSnapshotToolExecutorRunsSnapshot(t *testing.T) {
	service := &snapshotAnalystService{}
	executor := NewBusinessSnapshotToolExecutor(service)
	ctx := context.WithValue(context.Background(), authPkg.AuthUserKey, &authPkg.AuthUser{
		BusinessID: 11,
		UserID:     22,
	})

	output, executed, err := executor.ExecuteCaretTool(ctx, businessSnapshotToolName, map[string]any{
		"question": "How are we doing today?",
	})
	if err != nil {
		t.Fatalf("ExecuteCaretTool: %v", err)
	}
	if !executed {
		t.Fatal("expected business snapshot tool to execute")
	}
	if output["section_count"] != len(buildBusinessSnapshotQueries(timeNowUTCForTest())) {
		t.Fatalf("section_count = %#v", output["section_count"])
	}
	if output["truncated"] != false {
		t.Fatalf("truncated = %#v", output["truncated"])
	}
	sections, ok := output["sections"].([]map[string]any)
	if !ok || len(sections) == 0 {
		t.Fatalf("sections = %#v", output["sections"])
	}
	if len(service.calls) != len(buildBusinessSnapshotQueries(timeNowUTCForTest())) {
		t.Fatalf("expected all snapshot queries to run, got %d calls", len(service.calls))
	}
}

func TestBusinessSnapshotToolExecutorIgnoresOtherTools(t *testing.T) {
	executor := NewBusinessSnapshotToolExecutor(&snapshotAnalystService{})
	output, executed, err := executor.ExecuteCaretTool(context.Background(), "crm.list_customers", nil)
	if err != nil {
		t.Fatalf("ExecuteCaretTool: %v", err)
	}
	if executed || output != nil {
		t.Fatalf("expected other tool to be ignored, executed=%v output=%#v", executed, output)
	}
}

func timeNowUTCForTest() time.Time {
	return time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC)
}

type snapshotAnalystService struct {
	calls []string
}

func (s *snapshotAnalystService) SchemaPrompt() string {
	return "ai_pos_orders(total, payment_status, paid_at)\n"
}

func (s *snapshotAnalystService) RunJSON(_ context.Context, _, _ int64, specJSON string) ([]map[string]any, string, error) {
	s.calls = append(s.calls, specJSON)
	switch {
	case strings.Contains(specJSON, `"as":"revenue"`):
		return []map[string]any{{"revenue": 100}}, "SELECT", nil
	case strings.Contains(specJSON, `"as":"orders"`):
		return []map[string]any{{"orders": 4}}, "SELECT", nil
	case strings.Contains(specJSON, `"as":"amount_due"`):
		return []map[string]any{{"amount_due": 25, "invoice_count": 1}}, "SELECT", nil
	case strings.Contains(specJSON, `"as":"expenses"`):
		return []map[string]any{{"expenses": 10}}, "SELECT", nil
	case strings.Contains(specJSON, `"as":"customers"`):
		return []map[string]any{{"customers": 9}}, "SELECT", nil
	case strings.Contains(specJSON, `"as":"employees"`):
		return []map[string]any{{"employees": 2}}, "SELECT", nil
	case strings.Contains(specJSON, `"value_col":"low_stock_threshold"`):
		return []map[string]any{{
			"name":                "Whey",
			"sku":                 "WHEY-1",
			"stock_quantity":      0,
			"low_stock_threshold": 2,
		}}, "SELECT", nil
	case strings.Contains(specJSON, `"as":"tasks"`):
		return []map[string]any{{"status": "todo", "tasks": 3}}, "SELECT", nil
	default:
		return []map[string]any{{"ok": true}}, "SELECT", nil
	}
}
