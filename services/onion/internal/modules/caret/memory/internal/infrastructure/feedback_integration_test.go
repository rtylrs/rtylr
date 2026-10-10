package infrastructure

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
	"github.com/voxire/onion/internal/modules/organization/scopes"
	"github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/testutil"
)

func TestReplyFeedbackBecomesLessons(t *testing.T) {
	if host := os.Getenv("DB_HOST"); host != "" && host != "localhost" && host != "127.0.0.1" {
		t.Fatal("feedback fixtures require a local test database")
	}
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const userID, businessID = 741251, 910091
	suffix := uuid.NewString()
	if _, err := db.ExecContext(ctx, `INSERT INTO users (id,business_id,email,password_hash) VALUES (?, ?, ?, 'fixture')`, userID, businessID, suffix+"@caret.test"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, query := range []string{"DELETE FROM caret_reply_feedback WHERE user_id=?", "DELETE FROM user_memory WHERE user_id=?", "DELETE FROM users WHERE id=?"} {
			if _, err := db.ExecContext(ctx, query, userID); err != nil {
				t.Errorf("cleanup: %v", err)
			}
		}
	})
	repo := NewRepository(db).WithIdentityReaders(scopes.NewReader(db).DefaultBusiness, scopes.NewReader(db).AssistantIdentity)
	save := func(rating, reason string) {
		t.Helper()
		if err := repo.SaveReplyFeedback(ctx, userID, domain.ReplyFeedback{SessionID: "s-" + suffix, MessageID: 55, Rating: rating, Reason: reason, Prompt: "p", Reply: "r"}); err != nil {
			t.Fatalf("save %s: %v", rating, err)
		}
	}
	save("up", "")
	save("down", "too long") // replaces the thumbs up
	save("correction", "no, in USD")
	pending, err := repo.PendingReplyFeedback(ctx, userID, 20)
	if err != nil || len(pending) != 2 || pending[0].Rating != "down" || pending[0].Reason != "too long" || pending[1].Rating != "correction" {
		t.Fatalf("pending = %#v, %v", pending, err)
	}
	if err := repo.ReplaceLessons(ctx, userID, []string{"Keep replies short.", "Show money in USD."}, []int64{pending[0].ID, pending[1].ID}); err != nil {
		t.Fatal(err)
	}
	if left, _ := repo.PendingReplyFeedback(ctx, userID, 20); len(left) != 0 {
		t.Fatalf("feedback still pending after distillation: %#v", left)
	}
	lessons, err := repo.UserLessons(ctx, userID)
	if err != nil || len(lessons) != 2 || lessons[1] != "Show money in USD." {
		t.Fatalf("lessons = %q, %v", lessons, err)
	}
	// Lessons are standing preferences: they load into every chat here.
	authed := auth.WithAuthUser(ctx, &auth.AuthUser{UserID: userID, BusinessID: businessID})
	memories, err := repo.AssistantMemories(authed, "any-chat")
	if err != nil || len(memories) != 2 || memories[0].MemoryType != "lesson" {
		t.Fatalf("assistant memories = %#v, %v", memories, err)
	}
	if err := repo.ReplaceLessons(ctx, userID, []string{"Keep replies short."}, nil); err != nil {
		t.Fatal(err)
	}
	if lessons, _ := repo.UserLessons(ctx, userID); len(lessons) != 1 {
		t.Fatalf("a replaced set must drop removed lessons: %q", lessons)
	}
	if err := repo.SaveToolRoute(ctx, userID, "hr.list_leave_requests", "asked time off"); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveToolRoute(ctx, userID, "hr.list_leave_requests", "asked time off"); err != nil {
		t.Fatal(err)
	}
	if routes, err := repo.ToolRoutes(ctx, userID); err != nil || len(routes) != 1 || routes[0] != "hr.list_leave_requests\tasked time off" {
		t.Fatalf("routes = %q, %v", routes, err)
	}
	if listed, err := repo.ListMemories(authed); err != nil || len(listed) != 1 {
		t.Fatalf("a tool route must stay out of the memory list: %#v %v", listed, err)
	}
	save("clear", "")
	save("correction", "again")
	if pending, _ := repo.PendingReplyFeedback(ctx, userID, 20); len(pending) != 1 || pending[0].Reason != "again" {
		t.Fatalf("clear must remove the thumbs and a new correction must be pending: %#v", pending)
	}
}
