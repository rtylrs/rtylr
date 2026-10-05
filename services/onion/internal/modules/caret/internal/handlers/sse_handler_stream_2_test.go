package handlers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStreamRendererRedactsScreenshotLeakPattern(t *testing.T) {
	rec := httptest.NewRecorder()
	full := &strings.Builder{}
	renderer := newStreamEventRenderer(rec, nil, "assistant-1", full)

	text := "I obtained this information using the insights.get_customer_kpis and insights.get_product_kpis tools, which are part of the shortlisted capability catalog for the current request.\nNext Move: Run the insights.get_customer_kpis and insights.get_product_kpis tools to obtain the latest data."
	if err := renderer.AppendText(text); err != nil {
		t.Fatalf("append screenshot leak pattern: %v", err)
	}
	if err := renderer.Finish(); err != nil {
		t.Fatalf("finish stream: %v", err)
	}

	for _, banned := range []string{"insights.get_customer_kpis", "insights.get_product_kpis", "shortlisted capability catalog", "Next Move: Run"} {
		if strings.Contains(strings.ToLower(full.String()), strings.ToLower(banned)) {
			t.Fatalf("public response exposed %q: %q", banned, full.String())
		}
	}
}

func TestStreamRendererSilentlyDropsInvalidToolMarker(t *testing.T) {
	rec := httptest.NewRecorder()
	full := &strings.Builder{}
	renderer := newStreamEventRenderer(rec, nil, "assistant-1", full)

	err := renderer.AppendText(`Before <tool>{"input":{"name":"Ada"}}</tool> after.`)
	if err != nil {
		t.Fatalf("append tool chunk: %v", err)
	}
	if err := renderer.Finish(); err != nil {
		t.Fatalf("finish stream: %v", err)
	}

	events := decodeSSEEvents(t, rec.Body.String())
	for _, event := range events {
		if event.Type == EventError {
			t.Fatalf("invalid marker should not be surfaced as a UI error: %#v", events)
		}
		if event.Type == EventToolCall {
			t.Fatalf("invalid marker should not emit a tool call: %#v", events)
		}
	}
	if strings.Contains(full.String(), "<tool>") || strings.Contains(full.String(), "input") {
		t.Fatalf("public response should not include raw invalid tool marker, got %q", full.String())
	}
	if full.String() != "Before  after." {
		t.Fatalf("expected visible text around marker to remain, got %q", full.String())
	}
}

func TestStreamRendererSilentlyDropsReadToolMarker(t *testing.T) {
	rec := httptest.NewRecorder()
	full := &strings.Builder{}
	renderer := newStreamEventRenderer(rec, nil, "assistant-1", full)

	err := renderer.AppendText(`Done. <tool>{"name":"crm.list_customers","input":{}}</tool>`)
	if err != nil {
		t.Fatalf("append tool chunk: %v", err)
	}
	if err := renderer.Finish(); err != nil {
		t.Fatalf("finish stream: %v", err)
	}

	events := decodeSSEEvents(t, rec.Body.String())
	for _, event := range events {
		if event.Type == EventToolCall || event.Type == EventToolResult || event.Type == EventError {
			t.Fatalf("read tool marker should be silent, got %#v", events)
		}
	}
	if full.String() != "Done. " {
		t.Fatalf("expected text before read marker to remain, got %q", full.String())
	}
}

func TestStreamRendererSilentlyDropsUnhandledReadToolMarker(t *testing.T) {
	rec := httptest.NewRecorder()
	full := &strings.Builder{}
	renderer := newStreamEventRenderer(
		rec,
		nil,
		"assistant-1",
		full,
		withStreamToolExecution(context.Background(), []ToolExecutor{
			fakeToolExecutor{executed: false},
		}),
	)

	err := renderer.AppendText(`Done. <tool>{"name":"crm.list_customers","input":{}}</tool>`)
	if err != nil {
		t.Fatalf("append tool chunk: %v", err)
	}
	if err := renderer.Finish(); err != nil {
		t.Fatalf("finish stream: %v", err)
	}

	events := decodeSSEEvents(t, rec.Body.String())
	for _, event := range events {
		if event.Type == EventToolCall || event.Type == EventToolResult || event.Type == EventError {
			t.Fatalf("unhandled read tool marker should be silent, got %#v", events)
		}
	}
	if full.String() != "Done. " {
		t.Fatalf("expected text before read marker to remain, got %q", full.String())
	}
	if renderer.ToolResultCount() != 0 {
		t.Fatalf("unhandled read marker should not be collected as a tool result")
	}
}

func TestStreamRendererExecutesReadToolMarkerWithExecutor(t *testing.T) {
	rec := httptest.NewRecorder()
	full := &strings.Builder{}
	renderer := newStreamEventRenderer(
		rec,
		nil,
		"assistant-1",
		full,
		withStreamToolExecution(context.Background(), []ToolExecutor{
			fakeToolExecutor{output: map[string]any{
				"customers": []any{},
				"total":     0,
			}},
		}),
	)

	err := renderer.AppendText(`Done. <tool>{"name":"crm.list_customers","input":{"limit":2}}</tool>`)
	if err != nil {
		t.Fatalf("append tool chunk: %v", err)
	}
	if err := renderer.Finish(); err != nil {
		t.Fatalf("finish stream: %v", err)
	}

	events := decodeSSEEvents(t, rec.Body.String())
	var toolCall, toolResult *StreamEvent
	for i := range events {
		switch events[i].Type {
		case EventToolCall:
			toolCall = &events[i]
		case EventToolResult:
			toolResult = &events[i]
		}
	}
	if toolCall == nil {
		t.Fatalf("expected read tool_call event, got %#v", events)
	}
	if toolCall.Name != "crm.list_customers" || toolCall.RequiresApproval {
		t.Fatalf("expected executable read tool call, got %#v", *toolCall)
	}
	if toolResult == nil || toolResult.Status != "success" {
		t.Fatalf("expected successful tool result, got %#v", events)
	}
	if toolResult.Summary != "No customers found." {
		t.Fatalf("expected business summary, got %#v", toolResult.Summary)
	}
	output, ok := toolResult.Output.(map[string]any)
	if !ok || output["total"] != float64(0) {
		t.Fatalf("expected output total, got %#v", toolResult.Output)
	}
	if full.String() != "Done. " {
		t.Fatalf("expected visible text before marker to remain, got %q", full.String())
	}
	if renderer.ToolResultCount() != 1 {
		t.Fatalf("expected one collected tool result, got %d", renderer.ToolResultCount())
	}
	results := renderer.ToolResultsSince(0)
	if len(results) != 1 {
		t.Fatalf("expected one result since 0, got %#v", results)
	}
	if results[0].Name != "crm.list_customers" || results[0].Status != "success" {
		t.Fatalf("unexpected collected tool result: %#v", results[0])
	}
	if results[0].Summary != "No customers found." {
		t.Fatalf("expected collected summary, got %#v", results[0].Summary)
	}
	if results[0].Output["total"] != 0 {
		t.Fatalf("expected collected output total, got %#v", results[0].Output)
	}
}

type recordingRendererToolExecutor struct {
	input  map[string]any
	output map[string]any
}

func (e *recordingRendererToolExecutor) ExecuteCaretTool(_ context.Context, _ string, input map[string]any) (map[string]any, bool, error) {
	e.input = input
	return e.output, true, nil
}

func TestPlannerOnlyRendererSuppressesProviderProseAndBindsSelectedScope(t *testing.T) {
	rec := httptest.NewRecorder()
	full := &strings.Builder{}
	executor := &recordingRendererToolExecutor{output: map[string]any{"name": "Ada Secret"}}
	renderer := newStreamEventRenderer(
		rec,
		nil,
		"assistant-1",
		full,
		withStreamMode(ModeAsk),
		withStreamPlannerOnly(true),
		withStreamPageContext(map[string]any{
			"selected": map[string]any{"customer_id": "cus_server_scope", "name": "Ada Secret"},
		}),
		withStreamToolExecution(context.Background(), []ToolExecutor{executor}),
	)

	if err := renderer.AppendText(`<thinking>private provider reasoning</thinking>I already know Ada Secret. <tool>{"name":"crm.get_customer","input":{}}</tool>`); err != nil {
		t.Fatalf("append planner stream: %v", err)
	}
	if err := renderer.Finish(); err != nil {
		t.Fatalf("finish planner stream: %v", err)
	}

	if full.String() != "" {
		t.Fatalf("provider prose crossed the planner boundary: %q", full.String())
	}
	if strings.Contains(rec.Body.String(), "I already know Ada Secret") || strings.Contains(rec.Body.String(), "private provider reasoning") {
		t.Fatalf("provider prose crossed the SSE boundary: %s", rec.Body.String())
	}
	if executor.input["customer_id"] != "cus_server_scope" {
		t.Fatalf("selected scope was not bound server-side: %#v", executor.input)
	}
	events := decodeSSEEvents(t, rec.Body.String())
	var foundToolResult bool
	for _, event := range events {
		if event.Type == EventToolResult && event.Status == "success" {
			foundToolResult = true
		}
	}
	if !foundToolResult {
		t.Fatalf("expected server-rendered tool result, got %#v", events)
	}
}

func TestStreamRendererSilentlyDropsIncompleteToolMarker(t *testing.T) {
	rec := httptest.NewRecorder()
	full := &strings.Builder{}
	renderer := newStreamEventRenderer(rec, nil, "assistant-1", full)

	if err := renderer.AppendText(`Before <tool>{"name":"crm.create_customer"`); err != nil {
		t.Fatalf("append tool chunk: %v", err)
	}
	if err := renderer.Finish(); err != nil {
		t.Fatalf("finish stream: %v", err)
	}

	events := decodeSSEEvents(t, rec.Body.String())
	for _, event := range events {
		if event.Type == EventError || event.Type == EventToolCall {
			t.Fatalf("incomplete marker should be silent, got %#v", events)
		}
	}
	if full.String() != "Before " {
		t.Fatalf("expected text before incomplete marker to remain, got %q", full.String())
	}
}

func TestWriteAnalystToolEventsEmitsReadOnlyToolCallAndResult(t *testing.T) {
	handler := &GenerateTextStreamHandler{}
	rec := httptest.NewRecorder()
	grounding := &analystGroundingResult{
		Question:   "revenue today",
		SpecJSON:   `{"view":"ai_invoices","aggregations":[{"func":"sum","col":"total","as":"revenue"}]}`,
		Rows:       []map[string]any{{"revenue": 123.45}},
		RowCount:   1,
		Truncated:  false,
		DurationMs: 42,
	}

	if err := handler.writeAnalystToolEvents(rec, "assistant-1", grounding); err != nil {
		t.Fatalf("writeAnalystToolEvents: %v", err)
	}

	events := decodeSSEEvents(t, rec.Body.String())
	if len(events) != 2 {
		t.Fatalf("expected tool call and result, got %d events: %#v", len(events), events)
	}
	if events[0].Type != EventToolCall {
		t.Fatalf("first event should be tool_call, got %#v", events[0])
	}
	if events[0].Name != "analytics.query" {
		t.Fatalf("expected analytics.query tool, got %#v", events[0])
	}
	if events[0].RequiresApproval {
		t.Fatalf("read-only analytics tool should not require approval")
	}
	if events[1].Type != EventToolResult || events[1].Status != "success" {
		t.Fatalf("second event should be successful tool_result, got %#v", events[1])
	}
	output, ok := events[1].Output.(map[string]any)
	if !ok {
		t.Fatalf("expected output map, got %#v", events[1].Output)
	}
	if output["row_count"] != float64(1) {
		t.Fatalf("expected row_count output, got %#v", output)
	}
	if output["duration_ms"] != float64(42) {
		t.Fatalf("expected duration_ms output, got %#v", output)
	}
}

func TestWriteAnalystToolEventsUsesGroundingToolName(t *testing.T) {
	handler := &GenerateTextStreamHandler{}
	rec := httptest.NewRecorder()
	grounding := &analystGroundingResult{
		ToolName:   "business.snapshot",
		Question:   "how are we doing?",
		SpecJSON:   `{"kind":"business_snapshot"}`,
		Rows:       []map[string]any{{"metric": "today_revenue", "rows": []map[string]any{{"revenue": 10}}}},
		RowCount:   1,
		DurationMs: 4,
	}

	if err := handler.writeAnalystToolEvents(rec, "assistant-1", grounding); err != nil {
		t.Fatalf("writeAnalystToolEvents: %v", err)
	}

	events := decodeSSEEvents(t, rec.Body.String())
	if len(events) != 2 {
		t.Fatalf("expected tool call and result, got %#v", events)
	}
	if events[0].Name != "business.snapshot" {
		t.Fatalf("expected business.snapshot tool, got %#v", events[0])
	}
	if events[0].RequiresApproval {
		t.Fatalf("business snapshot should not require approval")
	}
}

func eventTypes(events []StreamEvent) []StreamEventType {
	out := make([]StreamEventType, len(events))
	for i := range events {
		out[i] = events[i].Type
	}
	return out
}

func streamEventTypeStrings(types []StreamEventType) []string {
	out := make([]string, len(types))
	for i := range types {
		out[i] = string(types[i])
	}
	return out
}

func decodeSSEEvents(t *testing.T, body string) []StreamEvent {
	t.Helper()
	lines := strings.Split(body, "\n")
	out := make([]StreamEvent, 0)
	for _, line := range lines {
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var ev StreamEvent
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &ev); err != nil {
			t.Fatalf("decode event %q: %v", line, err)
		}
		out = append(out, ev)
	}
	return out
}
