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

	echoapp "github.com/voxire/onion/internal/modules/echo/internal/application"
	"github.com/voxire/onion/internal/platform/aicache"
	"github.com/voxire/onion/internal/platform/snips"
)

func TestWriteSSEHeartbeatUsesCommentFrame(t *testing.T) {
	rec := httptest.NewRecorder()
	if err := WriteSSEHeartbeat(rec, nil); err != nil {
		t.Fatalf("WriteSSEHeartbeat: %v", err)
	}
	if got := rec.Body.String(); got != ": ping\n\n" {
		t.Fatalf("heartbeat = %q, want SSE comment frame", got)
	}
}

func TestDrainProviderStreamEmitsHeartbeatWhileProviderIsQuiet(t *testing.T) {
	handler := &GenerateTextStreamHandler{}
	rec := httptest.NewRecorder()
	full := &strings.Builder{}
	renderer := newStreamEventRenderer(rec, nil, "assistant-1", full)
	tokens := make(chan snips.StreamToken, 1)
	go func() {
		time.Sleep(10 * time.Millisecond)
		tokens <- snips.StreamToken{Done: true}
	}()

	if _, ok := handler.drainProviderStreamWithHeartbeat(
		context.Background(), rec, tokens, make(chan bool), renderer,
		cacheWritePlan{Mode: string(ModeAsk), StartedAt: time.Now()},
		1, 2, 0, time.Millisecond,
	); !ok {
		t.Fatal("drainProviderStreamWithHeartbeat did not complete")
	}
	if !strings.Contains(rec.Body.String(), ": ping\n\n") {
		t.Fatalf("stream body did not contain heartbeat: %q", rec.Body.String())
	}
}

func TestStreamDirectResponsePersistsBeforeDoneEvent(t *testing.T) {
	rec := httptest.NewRecorder()
	memory := &detachedMemoryFake{}
	memory.storeHook = func() {
		if strings.Contains(rec.Body.String(), `"type":"done"`) {
			t.Error("assistant response was persisted after the done event")
		}
	}
	handler := &GenerateTextStreamHandler{MemoryService: memory}

	handler.streamDirectResponse(context.Background(), rec, &GenerateTextStreamRequest{SessionID: "session-1", UserID: 7}, "answer")
	if memory.storeCalls != 1 {
		t.Fatalf("StoreMessage calls = %d, want 1", memory.storeCalls)
	}
	if !strings.Contains(rec.Body.String(), `"type":"done"`) {
		t.Fatalf("stream did not finish: %q", rec.Body.String())
	}
}

func TestCachedReplayDoesNotPersistAssistantHistory(t *testing.T) {
	memory := &detachedMemoryFake{}
	handler := &GenerateTextStreamHandler{MemoryService: memory}
	rec := httptest.NewRecorder()
	handler.streamCachedResponse(
		context.Background(), rec, make(chan bool),
		&GenerateTextStreamRequest{SessionID: "session-1", UserID: 7},
		aicache.ExactResponse{Response: "cached answer", Provider: "test"},
		cacheWritePlan{Key: "cache-key", Mode: string(ModeAsk)}, 1, 7,
	)
	if memory.storeCalls != 0 {
		t.Fatalf("cache replay StoreMessage calls = %d, want 0", memory.storeCalls)
	}
}

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

func TestCaretSmokeCheckPromptsAreHandledWithoutBusinessGeneration(t *testing.T) {
	for _, prompt := range []string{"test", " TEST! ", "ping", "hello", "is this working?"} {
		if !isCaretSmokeCheck(prompt) {
			t.Fatalf("expected %q to be recognized as a smoke check", prompt)
		}
	}

	for _, prompt := range []string{"test today's revenue", "hello, show my orders", "is this customer active?"} {
		if isCaretSmokeCheck(prompt) {
			t.Fatalf("expected business request %q to reach normal handling", prompt)
		}
	}
}

func TestValidateRejectsWhitespaceOnlyPrompts(t *testing.T) {
	handler := &GenerateTextStreamHandler{}
	if err := handler.Validate(context.Background(), &GenerateTextStreamRequest{Prompt: " \n\t ", UserID: 1}); err == nil {
		t.Fatal("expected whitespace-only prompt to be rejected")
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

func TestDrainProviderStreamReportsPrematureClosure(t *testing.T) {
	handler := &GenerateTextStreamHandler{}
	rec := httptest.NewRecorder()
	full := &strings.Builder{}
	renderer := newStreamEventRenderer(rec, nil, "assistant-1", full)
	tokens := make(chan snips.StreamToken)
	close(tokens)

	token, ok := handler.drainProviderStream(
		context.Background(),
		rec,
		tokens,
		make(chan bool),
		renderer,
		cacheWritePlan{Mode: string(ModeAsk), StartedAt: time.Now()},
		1,
		2,
		0,
	)
	if ok {
		t.Fatal("drainProviderStream succeeded after the provider stream closed")
	}
	if token.Error != unexpectedStreamClosureError {
		t.Fatalf("Error = %q, want %q", token.Error, unexpectedStreamClosureError)
	}

	events := decodeSSEEvents(t, rec.Body.String())
	if len(events) != 1 || events[0].Type != EventError || events[0].Error != unexpectedStreamClosureError {
		t.Fatalf("unexpected SSE events: %#v", events)
	}
}

func TestStreamTokensToClientRecordsPartialFailedUsage(t *testing.T) {
	quota := &recordingQuotaService{recorded: make(chan int, 1)}
	handler := &GenerateTextStreamHandler{QuotaService: quota}
	rec := httptest.NewRecorder()
	tokens := make(chan snips.StreamToken, 2)
	tokens <- snips.StreamToken{Text: "partial"}
	tokens <- snips.StreamToken{Error: "provider failed", Done: true}
	close(tokens)

	handler.streamTokensToClient(
		context.Background(),
		rec,
		tokens,
		make(chan bool),
		&GenerateTextStreamRequest{SessionID: "session-1", UserID: 2},
		1,
		2,
		cacheWritePlan{Mode: string(ModeAsk), PromptChars: 40, StartedAt: time.Now()},
		nil,
	)

	select {
	case got := <-quota.recorded:
		if got != estimateTokens(40)+estimateTokens(len("partial")) {
			t.Fatalf("recorded tokens = %d, want conservative partial usage", got)
		}
	case <-time.After(time.Second):
		t.Fatal("partial failed stream usage was not recorded")
	}
}

// failAfterMarkerWriter accepts writes until the recorded body contains marker,
// then fails every write. It models a client that disconnects right after the
// answer finished streaming but before the closing SSE events are written.
type failAfterMarkerWriter struct {
	*httptest.ResponseRecorder
	marker  string
	failing bool
}

func (w *failAfterMarkerWriter) Write(p []byte) (int, error) {
	if w.failing {
		return 0, errors.New("client disconnected")
	}
	n, err := w.ResponseRecorder.Write(p)
	if strings.Contains(w.Body.String(), w.marker) {
		w.failing = true
	}
	return n, err
}

// A client can disconnect after the provider stream completed and the answer
// was persisted but before the closing events are written. The tokens were
// consumed, so the deferred settle(0) must not refund the whole reservation.
func TestStreamTokensToClientSettlesObservedUsageWhenClosingWritesFail(t *testing.T) {
	quota := &recordingQuotaService{recorded: make(chan int, 1)}
	handler := &GenerateTextStreamHandler{
		QuotaService:  quota,
		MemoryService: &detachedMemoryFake{},
	}
	w := &failAfterMarkerWriter{
		ResponseRecorder: httptest.NewRecorder(),
		marker:           `"delta":"answer"`,
	}
	tokens := make(chan snips.StreamToken, 2)
	tokens <- snips.StreamToken{Text: "answer"}
	tokens <- snips.StreamToken{Done: true}
	close(tokens)

	plan := cacheWritePlan{Mode: string(ModeAsk), PromptChars: 40, StartedAt: time.Now()}
	observedTokens := streamUsageTokens(snips.StreamToken{}, plan.PromptChars, len("answer"))
	reservation := &quotaReservation{
		handler: handler, ctx: context.Background(), businessID: 1, userID: 2,
		reservedTokens: 4000,
	}

	func() {
		// Mirrors ServeHTTP's deferred protection against a leaked reservation.
		defer reservation.settle(0)
		handler.streamTokensToClient(
			context.Background(),
			w,
			tokens,
			make(chan bool),
			&GenerateTextStreamRequest{SessionID: "session-1", UserID: 2},
			1,
			2,
			plan,
			reservation,
		)
	}()

	select {
	case got := <-quota.recorded:
		want := int64(observedTokens) - reservation.reservedTokens
		if int64(got) != want {
			t.Fatalf("recorded delta = %d, want %d (observed %d, reserved %d)",
				got, want, observedTokens, reservation.reservedTokens)
		}
	case <-time.After(time.Second):
		t.Fatal("completed stream usage was not recorded")
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

type recordingQuotaService struct {
	recorded chan int
	reserved int64
}

func TestQuotaReservationSettlesOnceIncludingEarlyRelease(t *testing.T) {
	quota := &recordingQuotaService{recorded: make(chan int, 1)}
	handler := &GenerateTextStreamHandler{QuotaService: quota}
	reservation := &quotaReservation{
		handler: handler, ctx: context.Background(), businessID: 1, userID: 2,
		reservedTokens: 4000,
	}

	reservation.settle(0)
	reservation.settle(100)

	select {
	case got := <-quota.recorded:
		if got != -4000 {
			t.Fatalf("settlement delta = %d, want -4000", got)
		}
	case <-time.After(time.Second):
		t.Fatal("reservation was not settled")
	}
	select {
	case got := <-quota.recorded:
		t.Fatalf("reservation settled twice with delta %d", got)
	case <-time.After(50 * time.Millisecond):
	}
}

func (s *recordingQuotaService) ReserveQuota(_ context.Context, _, _ int64) (bool, string, int64, int64, int64, int64) {
	return true, "", 0, 0, 0, s.reserved
}

func (s *recordingQuotaService) ReconcileUsage(_ context.Context, _, _, actualTokens, reservedTokens int64) {
	s.recorded <- int(actualTokens - reservedTokens)
}

func TestBuildCacheWritePlanIgnoresRollingConversationMemory(t *testing.T) {
	now := time.Date(2026, 6, 19, 11, 0, 0, 0, time.UTC)
	req := &GenerateTextStreamRequest{
		Prompt:        "How does the exact cache work?",
		SystemContext: "User is in the Dash app.",
		Mode:          string(ModeAsk),
	}

	first := echoapp.BuildCacheWritePlan(
		echoapp.ModeAsk,
		echoapp.TurnRequest{Prompt: req.Prompt, SystemContext: req.SystemContext, PageContext: req.PageContext},
		"Recent messages (oldest first):\nuser: How does the exact cache work?\n\nCurrent request: How does the exact cache work?",
		11,
		22,
		"",
		now,
	)
	second := echoapp.BuildCacheWritePlan(
		echoapp.ModeAsk,
		echoapp.TurnRequest{Prompt: req.Prompt, SystemContext: req.SystemContext, PageContext: req.PageContext},
		"Recent messages (oldest first):\nuser: How does the exact cache work?\nassistant: cached answer\nuser: How does the exact cache work?\n\nCurrent request: How does the exact cache work?",
		11,
		22,
		"",
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
	other := echoapp.BuildCacheWritePlan(
		echoapp.ModeAsk,
		echoapp.TurnRequest{Prompt: otherReq.Prompt, SystemContext: otherReq.SystemContext, PageContext: otherReq.PageContext},
		"Current request: How does the exact cache work?",
		11,
		22,
		"",
		now,
	)
	if first.Key == other.Key {
		t.Fatal("different app context should not share an exact cache key")
	}
}

func TestBuildCacheWritePlanIncludesDataFreshness(t *testing.T) {
	now := time.Date(2026, 6, 20, 11, 0, 0, 0, time.UTC)
	req := &GenerateTextStreamRequest{
		Prompt:        "Which products are low?",
		SystemContext: "User is in inventory.",
		PageContext: map[string]any{
			"module":        "erp",
			"dataFreshness": "products:1",
		},
	}

	first := echoapp.BuildCacheWritePlan(echoapp.ModeAsk, echoapp.TurnRequest{Prompt: req.Prompt, SystemContext: req.SystemContext, PageContext: req.PageContext}, "Current request: Which products are low?", 11, 22, "", now)
	req.PageContext["dataFreshness"] = "products:2"
	second := echoapp.BuildCacheWritePlan(echoapp.ModeAsk, echoapp.TurnRequest{Prompt: req.Prompt, SystemContext: req.SystemContext, PageContext: req.PageContext}, "Current request: Which products are low?", 11, 22, "", now)

	if first.Enabled || second.Enabled || first.Key != "" || second.Key != "" {
		t.Fatalf("business-data turns must not be exact cached: first=%+v second=%+v", first, second)
	}
	if !first.ProviderPlannerOnly || !second.ProviderPlannerOnly {
		t.Fatalf("business-data turns must be planner-only: first=%+v second=%+v", first, second)
	}
}

func TestBuildCacheWritePlanIncludesGroundingFreshness(t *testing.T) {
	now := time.Date(2026, 6, 20, 11, 0, 0, 0, time.UTC)
	req := &GenerateTextStreamRequest{
		Prompt:        "How much revenue today?",
		SystemContext: "User is in dashboard.",
	}

	firstGrounding := &analystGroundingResult{
		ToolName: "analytics.query",
		SpecJSON: `{"view":"ai_pos_orders","aggregations":[{"func":"sum","col":"total","as":"revenue"}]}`,
		RowsJSON: `[{"revenue":100}]`,
		RowCount: 1,
	}
	secondGrounding := &analystGroundingResult{
		ToolName: "analytics.query",
		SpecJSON: firstGrounding.SpecJSON,
		RowsJSON: `[{"revenue":120}]`,
		RowCount: 1,
	}

	turnRequest := echoapp.TurnRequest{Prompt: req.Prompt, SystemContext: req.SystemContext, PageContext: req.PageContext}
	first := echoapp.BuildCacheWritePlan(echoapp.ModeAsk, turnRequest, "grounded prompt", 11, 22, "", now, firstGrounding)
	second := echoapp.BuildCacheWritePlan(echoapp.ModeAsk, turnRequest, "grounded prompt", 11, 22, "", now, secondGrounding)

	if first.Enabled || second.Enabled || first.Key != "" || second.Key != "" {
		t.Fatalf("grounded data turns must not be exact cached: first=%+v second=%+v", first, second)
	}
}

func TestShouldGroundBeforeExactCacheForLiveBusinessQuestions(t *testing.T) {
	cases := []string{
		"How much revenue today?",
		"top 5 customers",
		"show my top customers",
		"How many employees do I have?",
		"Which products are low on stock?",
	}
	for _, prompt := range cases {
		policy := echoapp.BuildCaretResponsePolicy(echoapp.ModeAsk, prompt, "", nil)
		if policy.GroundBeforeCache {
			t.Fatalf("provider path must not pre-ground %q", prompt)
		}
		if !policy.ProviderPlannerOnly {
			t.Fatalf("expected %q to use planner-only execution", prompt)
		}
	}

	if echoapp.BuildCaretResponsePolicy(echoapp.ModeSupport, "How much revenue today?", "", nil).GroundBeforeCache {
		t.Fatal("support mode should not use business grounding")
	}
	if echoapp.BuildCaretResponsePolicy(echoapp.ModeAsk, "How do I use Caret?", "", nil).GroundBeforeCache {
		t.Fatal("product help should not force live business grounding")
	}
}

func TestCachedReplayChunksPreserveTextAndSplitNaturally(t *testing.T) {
	text := "Here is the first cached sentence. Here is the second cached sentence, with enough words to replay in more than one visible delta. Final."

	chunks := cachedReplayChunks(text)
	if len(chunks) < 2 {
		t.Fatalf("expected cached replay to split text into visible chunks, got %d", len(chunks))
	}
	if strings.Join(chunks, "") != text {
		t.Fatalf("cached replay chunks changed text: %#v", chunks)
	}
	for i, chunk := range chunks[:len(chunks)-1] {
		if len([]rune(chunk)) > cachedReplayMaxChunkRunes {
			t.Fatalf("chunk %d too large: %q", i, chunk)
		}
	}
}

func TestCachedReplayChunksSplitShortAnswers(t *testing.T) {
	text := "You have 2 employees."

	chunks := cachedReplayChunks(text)
	if len(chunks) < 2 {
		t.Fatalf("expected short cached answer to replay visibly, got %d chunk", len(chunks))
	}
	if strings.Join(chunks, "") != text {
		t.Fatalf("cached replay chunks changed text: %#v", chunks)
	}
}

func TestCachedReplayDelayIsBounded(t *testing.T) {
	if got := cachedReplayDelay(1); got != 0 {
		t.Fatalf("single chunk delay = %s, want 0", got)
	}
	if got := cachedReplayDelay(3); got != cachedReplayMaxDelay {
		t.Fatalf("short replay delay = %s, want %s", got, cachedReplayMaxDelay)
	}
	if got := cachedReplayDelay(500); got != cachedReplayMinDelay {
		t.Fatalf("long replay delay = %s, want %s", got, cachedReplayMinDelay)
	}
}

func TestStreamRendererExtractsWriteToolProposal(t *testing.T) {
	rec := httptest.NewRecorder()
	full := &strings.Builder{}
	renderer := newStreamEventRenderer(rec, nil, "assistant-1", full, withStreamMode(ModeAgent))

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
	if toolCall.Mode != string(ModeAgent) {
		t.Fatalf("expected tool call to carry originating mode, got %#v", *toolCall)
	}
	if strings.Contains(full.String(), "<tool>") || strings.Contains(full.String(), "crm.create_customer") {
		t.Fatalf("public response should not include raw tool marker, got %q", full.String())
	}
}

func TestStreamRendererDropsWriteToolProposalOutsideAgentMode(t *testing.T) {
	rec := httptest.NewRecorder()
	full := &strings.Builder{}
	renderer := newStreamEventRenderer(rec, nil, "assistant-1", full, withStreamMode(ModeAsk))

	err := renderer.AppendText(`I can do that. <tool>{"name":"crm.create_customer","input":{"name":"Ada"}}</tool>`)
	if err != nil {
		t.Fatalf("append tool chunk: %v", err)
	}
	if err := renderer.Finish(); err != nil {
		t.Fatalf("finish stream: %v", err)
	}

	events := decodeSSEEvents(t, rec.Body.String())
	for _, event := range events {
		if event.Type == EventToolCall || event.Type == EventToolResult || event.Type == EventError {
			t.Fatalf("ask mode must not surface write tool markers, got %#v", events)
		}
	}
	if full.String() != "I can do that. " {
		t.Fatalf("expected visible text before marker to remain, got %q", full.String())
	}
}

func TestStreamRendererDoesNotExposeSplitInternalCapabilityName(t *testing.T) {
	rec := httptest.NewRecorder()
	full := &strings.Builder{}
	renderer := newStreamEventRenderer(rec, nil, "assistant-1", full)

	text := "I used insights.get_customer_kpis to check the customer data." + strings.Repeat(" ", markerHoldBytes+20)
	if err := renderer.AppendText(text); err != nil {
		t.Fatalf("append split internal detail: %v", err)
	}
	if err := renderer.Finish(); err != nil {
		t.Fatalf("finish stream: %v", err)
	}

	if strings.Contains(full.String(), "insights.get_customer_kpis") {
		t.Fatalf("public response exposed an internal capability name: %q", full.String())
	}
}

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
