package application

import (
	"context"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
)

type feedbackRepository struct {
	*chatRepository
	saved    []domain.ReplyFeedback
	lessons  []string
	marked   []int64
	existing []string
}

func (r *feedbackRepository) SaveReplyFeedback(_ context.Context, _ int, f domain.ReplyFeedback) error {
	r.saved = append(r.saved, f)
	return nil
}

func (r *feedbackRepository) PendingReplyFeedback(context.Context, int, int) ([]domain.ReplyFeedback, error) {
	return r.saved, nil
}

func (r *feedbackRepository) ReplaceLessons(_ context.Context, _ int, lessons []string, ids []int64) error {
	r.lessons, r.marked = lessons, ids
	return nil
}

func (r *feedbackRepository) UserLessons(context.Context, int) ([]string, error) {
	return r.existing, nil
}

func TestRecordReplyFeedbackReadsTheRatedExchangeFromTheChat(t *testing.T) {
	repo := &feedbackRepository{chatRepository: chatWithTurns(2, 5)}
	service := NewService(repo)
	pending, err := service.RecordReplyFeedback(context.Background(), "s", 1, "2", "down", "  too   long ")
	if err != nil || pending != 1 {
		t.Fatalf("pending=%d err=%v", pending, err)
	}
	got := repo.saved[0]
	if got.MessageID != 2 || !strings.HasPrefix(got.Prompt, "message-001") || !strings.HasPrefix(got.Reply, "message-002") || got.Reason != "too long" {
		t.Fatalf("feedback = %#v", got)
	}
	// A streaming id the server never stored rates the latest reply.
	if _, err := service.RecordReplyFeedback(context.Background(), "s", 1, "msg_abc", "up", ""); err != nil {
		t.Fatal(err)
	}
	if latest := repo.saved[1]; latest.MessageID != 4 || !strings.HasPrefix(latest.Prompt, "message-003") {
		t.Fatalf("latest = %#v", latest)
	}
	for _, bad := range [][2]string{{"meh", ""}, {"correction", ""}, {"down", strings.Repeat("x", 1001)}} {
		if _, err := service.RecordReplyFeedback(context.Background(), "s", 1, "", bad[0], bad[1]); err == nil {
			t.Errorf("accepted rating %q reason %d chars", bad[0], len(bad[1]))
		}
	}
}

func TestSaveLessonsDedupesAndCaps(t *testing.T) {
	repo := &feedbackRepository{chatRepository: chatWithTurns(1, 5)}
	lessons := []string{"Keep replies short.", "keep replies   short.", ""}
	for i := 0; i < 30; i++ {
		lessons = append(lessons, "Lesson "+strings.Repeat("x", i))
	}
	if err := NewService(repo).SaveLessons(context.Background(), 1, lessons, []int64{7}); err != nil {
		t.Fatal(err)
	}
	if len(repo.lessons) != MaxLessons || repo.lessons[0] != "Keep replies short." || repo.lessons[1] != "Lesson" || repo.marked[0] != 7 {
		t.Fatalf("lessons = %q marked = %v", repo.lessons, repo.marked)
	}
}
