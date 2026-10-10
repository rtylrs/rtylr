package handlers

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/snips"
	hubcaretv1 "github.com/voxire/voxie/gen/go/hub/caret/v1"
)

type recordingLearner struct {
	rating, reason string
	pending        int
	distilled      chan int
}

func (l *recordingLearner) RecordReplyFeedback(_ context.Context, _ string, _ int, _, rating, reason string) (int, error) {
	l.rating, l.reason = rating, reason
	return l.pending, nil
}

func (l *recordingLearner) PendingLessons(_ context.Context, userID int) ([]string, string, []int64, error) {
	l.distilled <- userID
	return nil, "", nil, nil
}

func (l *recordingLearner) SaveLessons(context.Context, int, []string, []int64) error { return nil }

type noopCompleter struct{}

func (noopCompleter) CompleteText(context.Context, string) (string, error) {
	return "<lessons></lessons>", nil
}

func TestRateReplyStoresTheVerdictAndLearnsFromAComplaint(t *testing.T) {
	learner := &recordingLearner{pending: 1, distilled: make(chan int, 1)}
	api := &RuntimeAPI{Learner: learner, AI: noopCompleter{}}
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{UserID: 12, BusinessID: 3})
	resp, err := api.RateReply(ctx, &hubcaretv1.RateReplyRequest{SessionId: "s1", Rating: "DOWN", Reason: "wrong currency"})
	if err != nil || !resp.GetSaved() {
		t.Fatalf("RateReply: %v %v", resp, err)
	}
	if learner.rating != "down" || learner.reason != "wrong currency" {
		t.Fatalf("stored %q %q", learner.rating, learner.reason)
	}
	select {
	case user := <-learner.distilled:
		if user != 12 {
			t.Fatalf("distilled for user %d", user)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("a thumbs down did not start a learning pass")
	}
	if _, err := api.RateReply(ctx, &hubcaretv1.RateReplyRequest{SessionId: "s1", Rating: "meh"}); err == nil {
		t.Fatal("an unknown rating was accepted")
	}
	if _, err := api.RateReply(context.Background(), &hubcaretv1.RateReplyRequest{SessionId: "s1", Rating: "up"}); err == nil {
		t.Fatal("an anonymous rating was accepted")
	}
}

type learningMemoryFake struct {
	detachedMemoryFake
	*recordingLearner
}

func TestCorrectionsInChatBecomeFeedbackOnThePreviousReply(t *testing.T) {
	learner := &recordingLearner{pending: 1, distilled: make(chan int, 1)}
	handler := &GenerateTextStreamHandler{AiService: &fakeAIStreamService{}, MemoryService: &learningMemoryFake{recordingLearner: learner}}
	handler.learnFromCorrection(context.Background(), "s1", 7, "No, I asked for last month")
	select {
	case user := <-learner.distilled:
		if user != 7 || learner.rating != "correction" || learner.reason != "No, I asked for last month" {
			t.Fatalf("user=%d rating=%q reason=%q", user, learner.rating, learner.reason)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("a correction did not start a learning pass")
	}
}

type storedReply int

func (s storedReply) StoredID() int { return int(s) }

type idMemoryFake struct{ detachedMemoryFake }

func (m *idMemoryFake) StoreMessage(context.Context, string, int, string, string, int) (any, error) {
	return storedReply(4321), nil
}

func TestMessageEndCarriesTheStoredReplyID(t *testing.T) {
	handler := &GenerateTextStreamHandler{AiService: &fakeAIStreamService{}, MemoryService: &idMemoryFake{}}
	rec := httptest.NewRecorder()
	tokens := make(chan snips.StreamToken, 2)
	tokens <- snips.StreamToken{Text: "Hello there."}
	tokens <- snips.StreamToken{Done: true}
	close(tokens)
	handler.streamTokensToClient(context.Background(), rec, tokens, make(chan bool),
		&GenerateTextStreamRequest{SessionID: "s1", UserID: 7}, 1, 7, cacheWritePlan{Mode: "agent"}, nil)
	if !strings.Contains(rec.Body.String(), `"type":"message_end"`) || !strings.Contains(rec.Body.String(), `"stored_id":"4321"`) {
		t.Fatalf("message_end lacks the stored id:\n%s", rec.Body.String())
	}
}

type recordingHistory struct{ notes []string }

func (h *recordingHistory) EnsureSessionOwnership(context.Context, string, int) error { return nil }

func (h *recordingHistory) StoreMessage(_ context.Context, _ string, _ int, role, content string, _ int) (any, error) {
	h.notes = append(h.notes, role+": "+content)
	return nil, nil
}

func TestDecisionsAreRecordedInTheirChat(t *testing.T) {
	history := &recordingHistory{}
	api := &RuntimeAPI{History: history}
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{UserID: 3, BusinessID: 1})
	api.recordDecision(ctx, "s1", ToolDecisionResponse{Status: "success", Executed: true, Preview: "Create task \"Call Acme\" due Friday", Message: "Task created."})
	api.recordDecision(ctx, "s1", ToolDecisionResponse{Status: "denied", Preview: "Delete product Latte"})
	api.recordDecision(ctx, "", ToolDecisionResponse{Status: "success", Executed: true, Preview: "x"})
	api.recordDecision(ctx, "s1", ToolDecisionResponse{Status: "error", Preview: "y"})
	if len(history.notes) != 2 || history.notes[0] != `assistant: Done: Create task "Call Acme" due Friday. Task created.` || history.notes[1] != "assistant: Skipped at your request: Delete product Latte. No changes were made." {
		t.Fatalf("notes = %q", history.notes)
	}
}
