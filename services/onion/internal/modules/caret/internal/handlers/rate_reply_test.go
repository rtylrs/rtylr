package handlers

import (
	"context"
	"testing"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
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

func (noopCompleter) CompleteText(context.Context, string) (string, error) { return "<lessons></lessons>", nil }

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
