package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
	for _, answer := range []string{"answer", caretSmokeCheckReply, strings.Repeat("A longer reply. ", 20)} {
		t.Run(answer[:6], func(t *testing.T) {
			rec := httptest.NewRecorder()
			memory := &detachedMemoryFake{}
			memory.storeHook = func() {
				if strings.Contains(rec.Body.String(), `"type":"done"`) {
					t.Error("assistant response was persisted after the done event")
				}
				if memory.storedContent != answer {
					t.Errorf("persisted response = %q, want complete answer %q", memory.storedContent, answer)
				}
			}
			handler := &GenerateTextStreamHandler{MemoryService: memory}

			handler.streamDirectResponse(context.Background(), rec, &GenerateTextStreamRequest{SessionID: "session-1", UserID: 7}, answer)
			if memory.storeCalls != 1 {
				t.Fatalf("StoreMessage calls = %d, want 1", memory.storeCalls)
			}
			if !strings.Contains(rec.Body.String(), `"type":"done"`) {
				t.Fatalf("stream did not finish: %q", rec.Body.String())
			}
		})
	}
}

func TestStreamDirectResponseRemovesEmDashesBeforeDisplayAndPersistence(t *testing.T) {
	memory := &detachedMemoryFake{}
	handler := &GenerateTextStreamHandler{MemoryService: memory}
	rec := httptest.NewRecorder()
	handler.streamDirectResponse(context.Background(), rec, &GenerateTextStreamRequest{SessionID: "session-1", UserID: 7}, "I checked—your tasks are on track.")

	if strings.ContainsRune(rec.Body.String(), '—') || strings.ContainsRune(memory.storedContent, '—') {
		t.Fatalf("em dash crossed the reply boundary: body=%q stored=%q", rec.Body.String(), memory.storedContent)
	}
	if memory.storedContent != "I checked, your tasks are on track." || !strings.Contains(rec.Body.String(), "I checked, your tasks are on track.") {
		t.Fatalf("displayed and persisted replies did not use the normalized text: body=%q stored=%q", rec.Body.String(), memory.storedContent)
	}
}

func TestCachedReplayPersistsBeforeCompletion(t *testing.T) {
	memory := &detachedMemoryFake{}
	handler := &GenerateTextStreamHandler{MemoryService: memory}
	rec := httptest.NewRecorder()
	memory.storeHook = func() {
		if strings.Contains(rec.Body.String(), `"type":"done"`) {
			t.Fatal("cache reply completed before persistence")
		}
	}
	handler.streamCachedResponse(
		context.Background(), rec, make(chan bool),
		&GenerateTextStreamRequest{SessionID: "session-1", UserID: 7},
		aicache.ExactResponse{Response: "cached answer", Provider: "test"},
		cacheWritePlan{Key: "cache-key", Mode: string(ModeAsk)}, 1, 7,
	)
	if memory.storeCalls != 1 || memory.storedContent != "cached answer" {
		t.Fatalf("cache replay did not store its answer: calls=%d content=%q", memory.storeCalls, memory.storedContent)
	}
}

func TestGroundedSearchIsStreamedWithItsSuggestionsButNotStoredAsChatAnswer(t *testing.T) {
	memory := &detachedMemoryFake{}
	handler := &GenerateTextStreamHandler{
		MemoryService: memory,
		ToolExecutors: []ToolExecutor{fakeToolExecutor{
			executed: true,
			output: map[string]any{
				"answer":                  "A grounded answer — with checked sources.",
				"sources":                 []map[string]string{{"title": "Source", "url": "https://example.com/source"}},
				"search_suggestions_html": `<div><a href="https://www.google.com/search?q=next">Related search</a></div>`,
			},
		}},
	}
	rec := httptest.NewRecorder()
	tokens := make(chan snips.StreamToken, 2)
	tokens <- snips.StreamToken{Text: `<tool>{"name":"web.search","input":{"query":"public query"}}</tool>`}
	tokens <- snips.StreamToken{Done: true}
	close(tokens)
	handler.streamTokensToClient(
		context.Background(), rec, tokens, make(chan bool),
		&GenerateTextStreamRequest{SessionID: "session-1", UserID: 7, EnhancedPrompt: "search the web"},
		1, 7, cacheWritePlan{Mode: string(ModeAgent), ContinuationBudget: 2}, nil,
	)

	body := rec.Body.String()
	if memory.storeCalls != 0 {
		t.Fatalf("grounded web result should not be persisted without its associated search suggestions; stored %d responses", memory.storeCalls)
	}
	for _, want := range []string{"A grounded answer, with checked sources.", "Related search", `"type":"done"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("stream missing %q: %s", want, body)
		}
	}
	if strings.ContainsRune(body, '—') {
		t.Fatalf("grounded web reply still contains an em dash: %s", body)
	}
}

func TestFailedWebSearchReturnsQuotaExplanationWithoutAnotherModelCall(t *testing.T) {
	memory := &detachedMemoryFake{}
	handler := &GenerateTextStreamHandler{
		MemoryService: memory,
		ToolExecutors: []ToolExecutor{fakeToolExecutor{err: errors.New("web search returned status 429")}},
	}
	rec := httptest.NewRecorder()
	tokens := make(chan snips.StreamToken, 2)
	tokens <- snips.StreamToken{Text: `<tool>{"name":"web.search","input":{"query":"recent public news"}}</tool>`}
	tokens <- snips.StreamToken{Done: true}
	close(tokens)
	handler.streamTokensToClient(
		context.Background(), rec, tokens, make(chan bool),
		&GenerateTextStreamRequest{SessionID: "session-1", UserID: 7, EnhancedPrompt: "search the web"},
		1, 7, cacheWritePlan{Mode: string(ModeAgent), ContinuationBudget: 3}, nil,
	)

	body := rec.Body.String()
	if !strings.Contains(body, "Web search hit Google's current request limit") || !strings.Contains(body, `"type":"done"`) {
		t.Fatalf("failed web search did not get a clear completed response: %s", body)
	}
	if memory.storeCalls != 1 || !strings.Contains(memory.storedContent, "current request limit") {
		t.Fatalf("quota explanation was not persisted as the assistant reply: calls=%d content=%q", memory.storeCalls, memory.storedContent)
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

type failingReplayMemory struct{ detachedMemoryFake }

func (m *failingReplayMemory) StoreMessage(context.Context, string, int, string, string, int) (any, error) {
	return nil, errors.New("fixture persistence failure")
}
func TestCachedReplayReportsPersistenceFailureWithoutCompletion(t *testing.T) {
	h := &GenerateTextStreamHandler{MemoryService: &failingReplayMemory{}}
	rec := httptest.NewRecorder()
	h.streamCachedResponse(context.Background(), rec, make(chan bool), &GenerateTextStreamRequest{SessionID: "fixture", UserID: 7}, aicache.ExactResponse{Response: "saved reply"}, cacheWritePlan{}, 1, 7)
	if !strings.Contains(rec.Body.String(), "Failed to persist response") || strings.Contains(rec.Body.String(), `"type":"done"`) {
		t.Fatalf("invalid persistence acknowledgement: %s", rec.Body.String())
	}
}

func TestEmptyAssistantResponseIsRetryableAndNotPersisted(t *testing.T) {
	memory := &detachedMemoryFake{}
	h := &GenerateTextStreamHandler{MemoryService: memory}
	tokens := make(chan snips.StreamToken, 1)
	tokens <- snips.StreamToken{Done: true}
	rec := httptest.NewRecorder()

	h.streamTokensToClient(context.Background(), rec, tokens, make(chan bool), &GenerateTextStreamRequest{
		SessionID: "session-1", UserID: 7,
	}, 1, 7, cacheWritePlan{Mode: string(ModeAgent), StartedAt: time.Now()}, nil)

	body := rec.Body.String()
	if !strings.Contains(body, `"type":"error"`) || !strings.Contains(body, "Please try again") {
		t.Fatalf("empty completion should produce a retryable stream error, got %q", body)
	}
	if strings.Contains(body, "reliable answer from the available business context") {
		t.Fatalf("empty completion used the misleading business-context fallback: %q", body)
	}
	if memory.storeCalls != 0 {
		t.Fatalf("empty completion persisted a fallback response: %#v", memory.storedContent)
	}
}
