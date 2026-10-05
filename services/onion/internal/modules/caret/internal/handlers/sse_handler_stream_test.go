package handlers

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	echoapp "github.com/voxire/onion/internal/modules/caret/internal/application"
	"github.com/voxire/onion/internal/platform/snips"
)

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
