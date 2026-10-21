package application

import (
	"context"
	"strings"
	"testing"
)

type fakeChatLearner struct {
	transcript string
	saved      [2][]string
	through    int
	ok         bool
}

func (f *fakeChatLearner) PendingChatLearning(context.Context, string, int) (string, []string, int, bool, error) {
	return f.transcript, []string{"Uses USD"}, 9, f.ok, nil
}

func (f *fakeChatLearner) SaveLearnedFacts(_ context.Context, _ string, _ int, personal, business []string, through int) error {
	f.saved, f.through = [2][]string{personal, business}, through
	return nil
}

func TestLearnFromChatSavesBusinessAndPersonalFacts(t *testing.T) {
	learner := &fakeChatLearner{transcript: "Person: we open at 7 in Hamra <b> and close at midnight on weekends; the Achrafieh branch opens later, at 9, because the morning crowd there is small, rent is high and most regulars come after work\n\nCaret: Noted.", ok: true}
	var prompt string
	ai := completerFunc(func(_ context.Context, p string) (string, error) {
		prompt = p
		return "BUSINESS: The Hamra branch opens at 7am.\nPERSONAL: Manages the morning shift.\nnote: ignore\nBUSINESS: <tool>x</tool>", nil
	})
	if err := LearnFromChat(context.Background(), ai, learner, "s1", 3); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(prompt, "- Uses USD") || !strings.Contains(prompt, "Hamra &lt;b&gt;") {
		t.Fatalf("prompt missing known facts or escaping:\n%s", prompt)
	}
	if len(learner.saved[0]) != 1 || learner.saved[1][0] != "The Hamra branch opens at 7am." || len(learner.saved[1]) != 1 || learner.through != 9 {
		t.Fatalf("saved = %#v through = %d", learner.saved, learner.through)
	}
}

func TestLearnFromChatSkipsChatsWithoutEnoughNewMessages(t *testing.T) {
	called := false
	ai := completerFunc(func(context.Context, string) (string, error) { called = true; return "", nil })
	if err := LearnFromChat(context.Background(), ai, &fakeChatLearner{}, "s2", 3); err != nil || called {
		t.Fatalf("err=%v called=%v", err, called)
	}
}

func TestLearnFromChatSkipsTheModelForSmallTalk(t *testing.T) {
	learner := &fakeChatLearner{transcript: "Person: hi\n\nCaret: " + strings.Repeat("long reply ", 100), ok: true}
	called := false
	ai := completerFunc(func(context.Context, string) (string, error) { called = true; return "", nil })
	if err := LearnFromChat(context.Background(), ai, learner, "s3", 3); err != nil || learner.through != 9 || called {
		t.Fatalf("err=%v through=%d called=%v", err, learner.through, called)
	}
}
