package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGenerateTextStreamParseRejectsTrailingJSON(t *testing.T) {
	handler := &GenerateTextStreamHandler{}
	req := httptest.NewRequest(
		http.MethodPost,
		"/ai/generate-text-stream",
		strings.NewReader(`{"prompt":"hello","user_id":1}{"prompt":"ignored","user_id":2}`),
	)
	rec := httptest.NewRecorder()

	got, err := handler.parseAndValidateRequest(rec, req)

	if !errors.Is(err, ErrInvalidRequestBody) {
		t.Fatalf("expected invalid request body, got req=%v err=%v", got, err)
	}
}

func TestGenerateTextStreamParseAcceptsCamelCaseAliases(t *testing.T) {
	handler := &GenerateTextStreamHandler{}
	req := httptest.NewRequest(
		http.MethodPost,
		"/ai/generate-text-stream",
		strings.NewReader(`{"prompt":"hello","sessionId":"session-1","userId":7}`),
	)
	rec := httptest.NewRecorder()

	got, err := handler.parseAndValidateRequest(rec, req)

	if err != nil {
		t.Fatalf("parseAndValidateRequest: %v", err)
	}
	if got.SessionID != "session-1" {
		t.Fatalf("expected session alias to map, got %q", got.SessionID)
	}
	if got.UserID != 7 {
		t.Fatalf("expected user alias to map, got %d", got.UserID)
	}
}

func TestStreamPreludeAndTokensUseStructuredSSEEvents(t *testing.T) {
	handler := &GenerateTextStreamHandler{}
	rec := httptest.NewRecorder()
	full := &strings.Builder{}

	if err := handler.writeStreamPrelude(rec, "session-1", "assistant-1"); err != nil {
		t.Fatalf("writeStreamPrelude: %v", err)
	}
	renderer := newStreamEventRenderer(rec, nil, "assistant-1", full)
	if err := renderer.AppendText("hello"); err != nil {
		t.Fatalf("append text: %v", err)
	}
	if err := renderer.Finish(); err != nil {
		t.Fatalf("finish stream: %v", err)
	}

	events := decodeSSEEvents(t, rec.Body.String())
	wantTypes := []StreamEventType{
		EventSession,
		EventMessageStart,
		EventTextDelta,
		EventMessageEnd,
		EventDone,
	}
	if len(events) != len(wantTypes) {
		t.Fatalf("expected %d events, got %d: %#v", len(wantTypes), len(events), events)
	}
	for i, want := range wantTypes {
		if events[i].Type != want {
			t.Fatalf("event %d type: want %q got %q", i, want, events[i].Type)
		}
	}
	if events[0].SessionID != "session-1" {
		t.Fatalf("expected session event to carry session id, got %#v", events[0])
	}
	if events[1].Role != "assistant" {
		t.Fatalf("expected assistant role on message_start, got %#v", events[1])
	}
	if events[2].Delta != "hello" {
		t.Fatalf("expected text delta, got %#v", events[2])
	}
	if full.String() != "hello" {
		t.Fatalf("expected response accumulator to contain text, got %q", full.String())
	}
}

func TestStreamRendererExtractsSplitThinkingMarker(t *testing.T) {
	rec := httptest.NewRecorder()
	full := &strings.Builder{}
	renderer := newStreamEventRenderer(rec, nil, "assistant-1", full)

	if err := renderer.AppendText("Here is the answer. <think"); err != nil {
		t.Fatalf("append first chunk: %v", err)
	}
	if err := renderer.AppendText("ing>Checking live data</thinking> Final."); err != nil {
		t.Fatalf("append second chunk: %v", err)
	}
	if err := renderer.Finish(); err != nil {
		t.Fatalf("finish stream: %v", err)
	}

	events := decodeSSEEvents(t, rec.Body.String())
	types := streamEventTypeStrings(eventTypes(events))
	for _, want := range []string{"thinking_start", "thinking_delta", "thinking_end", "message_end", "done"} {
		if !strings.Contains(strings.Join(types, ","), want) {
			t.Fatalf("missing %s in event types %#v\nbody:\n%s", want, types, rec.Body.String())
		}
	}
	if strings.Contains(full.String(), "<thinking>") || strings.Contains(full.String(), "Checking live data") {
		t.Fatalf("public response should not persist thinking marker/content, got %q", full.String())
	}
	if !strings.Contains(full.String(), "Here is the answer.") || !strings.Contains(full.String(), "Final.") {
		t.Fatalf("public response missing visible text, got %q", full.String())
	}
}

func TestBuildCacheWritePlanIgnoresRollingConversationMemory(t *testing.T) {
	handler := &GenerateTextStreamHandler{}
	now := time.Date(2026, 6, 19, 11, 0, 0, 0, time.UTC)
	req := &GenerateTextStreamRequest{
		Prompt:        "Show my top customers",
		SystemContext: "User is in the Dash app.",
		Mode:          string(ModeAsk),
	}

	first := handler.buildCacheWritePlan(
		context.Background(),
		ModeAsk,
		req,
		"Recent messages (oldest first):\nuser: Show my top customers\n\nCurrent request: Show my top customers",
		11,
		22,
		now,
	)
	second := handler.buildCacheWritePlan(
		context.Background(),
		ModeAsk,
		req,
		"Recent messages (oldest first):\nuser: Show my top customers\nassistant: cached answer\nuser: Show my top customers\n\nCurrent request: Show my top customers",
		11,
		22,
		now,
	)

	if first.Key == "" || second.Key == "" {
		t.Fatalf("expected cache keys, got first=%q second=%q", first.Key, second.Key)
	}
	if first.Key != second.Key {
		t.Fatalf("same request identity should reuse cache key despite rolling memory: %q != %q", first.Key, second.Key)
	}
	if first.PromptChars == len(req.Prompt) {
		t.Fatal("prompt chars should still track the full provider prompt for quota estimates")
	}

	otherReq := *req
	otherReq.SystemContext = "User is in the POS app."
	other := handler.buildCacheWritePlan(
		context.Background(),
		ModeAsk,
		&otherReq,
		"Current request: Show my top customers",
		11,
		22,
		now,
	)
	if first.Key == other.Key {
		t.Fatal("different app context should not share an exact cache key")
	}
}

func TestStreamRendererExtractsWriteToolProposal(t *testing.T) {
	rec := httptest.NewRecorder()
	full := &strings.Builder{}
	renderer := newStreamEventRenderer(rec, nil, "assistant-1", full)

	err := renderer.AppendText(`I can do that. <tool>{"name":"crm.create_customer","input":{"name":"Ada"}}</tool> Waiting for approval.`)
	if err != nil {
		t.Fatalf("append tool chunk: %v", err)
	}
	if err := renderer.Finish(); err != nil {
		t.Fatalf("finish stream: %v", err)
	}

	events := decodeSSEEvents(t, rec.Body.String())
	var toolCall *StreamEvent
	for i := range events {
		if events[i].Type == EventToolCall {
			toolCall = &events[i]
			break
		}
	}
	if toolCall == nil {
		t.Fatalf("expected tool_call event, got %#v", events)
	}
	if toolCall.Name != "crm.create_customer" || !toolCall.RequiresApproval {
		t.Fatalf("expected approval-backed create customer proposal, got %#v", *toolCall)
	}
	if strings.Contains(full.String(), "<tool>") || strings.Contains(full.String(), "crm.create_customer") {
		t.Fatalf("public response should not include raw tool marker, got %q", full.String())
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
	output, ok := toolResult.Output.(map[string]any)
	if !ok || output["total"] != float64(0) {
		t.Fatalf("expected output total, got %#v", toolResult.Output)
	}
	if full.String() != "Done. " {
		t.Fatalf("expected visible text before marker to remain, got %q", full.String())
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
