package memory

import (
	"context"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/testutil"
)

func TestBuddyThreadIsStableAndHiddenFromTheChatList(t *testing.T) {
	db := testutil.OpenDB(t)
	const userID, businessID = 930711, int64(920504)
	clean := func() {
		for _, q := range []string{
			"DELETE FROM conversation_messages WHERE user_id = ?",
			"DELETE FROM chat_sessions WHERE user_id = ?",
		} {
			if _, err := db.Exec(q, userID); err != nil {
				t.Fatal(err)
			}
		}
	}
	clean()
	t.Cleanup(clean)
	m := NewModule(context.Background(), db).WithIdentityReaders(
		func(context.Context, int64) (int64, error) { return businessID, nil }, nil)
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{UserID: userID, BusinessID: businessID})

	first, err := m.BuddyThread(ctx, businessID, userID)
	if err != nil {
		t.Fatal(err)
	}
	again, err := m.BuddyThread(ctx, businessID, userID)
	if err != nil || again != first || first != "buddy-920504-930711" {
		t.Fatalf("thread ids %q %q: %v", first, again, err)
	}
	for _, role := range []string{"user", "assistant", "user"} {
		if _, err := m.StoreMessage(ctx, first, userID, role, role+" says hi", 3); err != nil {
			t.Fatal(err)
		}
	}
	recent, err := m.RecentMessages(ctx, first, userID, 2)
	if err != nil || len(recent) != 2 || recent[0].Role != "assistant" || recent[1].Role != "user" {
		t.Fatalf("recent = %+v, %v", recent, err)
	}
	if _, err := m.RecentMessages(ctx, first, userID+1, 2); err == nil {
		t.Fatal("another person read the buddy chat")
	}

	other, err := m.CreateSession(ctx, userID, "Normal chat")
	if err != nil {
		t.Fatal(err)
	}
	sessions, err := m.memoryService.ListUserSessions(ctx, userID, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].SessionID != other {
		t.Fatalf("chat list = %+v, want only the normal chat", sessions)
	}
}
