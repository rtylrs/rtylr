package handlers

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/caret/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/snips"
	hubcaretv1 "github.com/voxire/voxie/gen/go/hub/caret/v1"
)

// slowAI streams a reply in two parts with a pause between them.
type slowAI struct{ release chan struct{} }

func (s *slowAI) IsAvailable() bool          { return true }
func (s *slowAI) GetPrimaryProvider() string { return "test" }
func (s *slowAI) GenerateTextStream(ctx context.Context, _ string) (<-chan snips.StreamToken, error) {
	tokens := make(chan snips.StreamToken)
	go func() {
		defer close(tokens)
		tokens <- snips.StreamToken{Text: "First half. "}
		select {
		case <-s.release:
		case <-ctx.Done():
			return
		}
		tokens <- snips.StreamToken{Text: "Second half."}
		tokens <- snips.StreamToken{Done: true}
	}()
	return tokens, nil
}

// A reply whose client leaves keeps generating and is saved in full.
func TestReplyFinishesAndIsSavedAfterTheClientLeaves(t *testing.T) {
	ai := &slowAI{release: make(chan struct{})}
	memory := &detachedMemoryFake{}
	handler := &GenerateTextStreamHandler{AiService: ai, MemoryService: memory}
	tokens, _ := ai.GenerateTextStream(context.Background(), "")
	clientGone := make(chan bool, 1)
	w := newResilientWriter(httptest.NewRecorder())

	go func() {
		clientGone <- true
		close(ai.release)
	}()
	handler.streamTokensToClient(context.Background(), w, tokens, clientGone,
		&GenerateTextStreamRequest{SessionID: "s", UserID: 7}, 1, 7,
		cacheWritePlan{Mode: string(application.ModeAgent)}, nil)

	if memory.storeCalls != 1 || !strings.Contains(memory.storedContent, "Second half.") {
		t.Fatalf("reply not finished in the background: calls=%d content=%q", memory.storeCalls, memory.storedContent)
	}
}

func TestStopReplyCancelsOnlyTheCallersChat(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer activeReplies.start(7, "chat-1", cancel)()

	api := &RuntimeAPI{}
	other := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{UserID: 8, BusinessID: 1})
	if res, err := api.StopReply(other, &hubcaretv1.StopReplyRequest{SessionId: "chat-1"}); err != nil || res.GetStopped() {
		t.Fatalf("another user must not stop this reply: %+v %v", res, err)
	}
	owner := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{UserID: 7, BusinessID: 1})
	if res, err := api.StopReply(owner, &hubcaretv1.StopReplyRequest{SessionId: "chat-1"}); err != nil || !res.GetStopped() {
		t.Fatalf("owner stop failed: %+v %v", res, err)
	}
	if ctx.Err() == nil {
		t.Fatal("stop did not cancel the reply")
	}
}
