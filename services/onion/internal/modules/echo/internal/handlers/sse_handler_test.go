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
