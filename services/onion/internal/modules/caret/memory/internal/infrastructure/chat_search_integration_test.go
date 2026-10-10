package infrastructure

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/modules/organization/scopes"
	"github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/testutil"
)

func TestSearchChatsStaysInTheUsersBusiness(t *testing.T) {
	if host := os.Getenv("DB_HOST"); host != "" && host != "localhost" && host != "127.0.0.1" {
		t.Fatal("chat fixtures require a local test database")
	}
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const userID, otherUserID, businessID, otherBusinessID = 741241, 741242, 910081, 910082
	suffix := uuid.NewString()
	if _, err := db.ExecContext(ctx, `INSERT INTO users (id,business_id,email,password_hash) VALUES (?, ?, ?, 'fixture'), (?, ?, ?, 'fixture')`,
		userID, businessID, suffix+"@caret.test", otherUserID, businessID, "other-"+suffix+"@caret.test"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, query := range []string{
			"DELETE FROM conversation_messages WHERE user_id IN (?,?)",
			"DELETE FROM chat_sessions WHERE user_id IN (?,?)",
			"DELETE FROM users WHERE id IN (?,?)",
		} {
			if _, err := db.ExecContext(ctx, query, userID, otherUserID); err != nil {
				t.Errorf("remove chat fixtures: %v", err)
			}
		}
	})
	here, there, others := "here-"+suffix, "there-"+suffix, "others-"+suffix
	if _, err := db.ExecContext(ctx, `INSERT INTO chat_sessions (user_id,session_id,title,metadata) VALUES
		(?,?,'Pricing',JSON_OBJECT('business_id',?)),(?,?,'Elsewhere',JSON_OBJECT('business_id',?)),(?,?,'Not mine',JSON_OBJECT('business_id',?))`,
		userID, here, businessID, userID, there, otherBusinessID, otherUserID, others, businessID); err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		user          int
		session, text string
	}{
		{userID, here, "We decided the espresso price goes to 4 dollars next month"},
		{userID, there, "The espresso supplier for the other shop is Kahwa"},
		{otherUserID, others, "My espresso notes are private"},
	} {
		if _, err := db.ExecContext(ctx, `INSERT INTO conversation_messages (session_id,user_id,role,content) VALUES (?,?,'user',?)`, row.session, row.user, row.text); err != nil {
			t.Fatal(err)
		}
	}
	repo := NewRepository(db).WithIdentityReaders(scopes.NewReader(db).DefaultBusiness, scopes.NewReader(db).AssistantIdentity)
	ctx = auth.WithAuthUser(ctx, &auth.AuthUser{UserID: userID, BusinessID: businessID})
	if err := repo.SetChatMetadataNumber(ctx, here, userID, "learned_through_id", 42); err != nil {
		t.Fatal(err)
	}
	var metadata string
	if err := db.QueryRowContext(ctx, `SELECT metadata FROM chat_sessions WHERE session_id=?`, here).Scan(&metadata); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(metadata, `"learned_through_id": 42`) || !strings.Contains(metadata, `"business_id": 910081`) {
		t.Fatalf("metadata = %s", metadata)
	}
	if err := repo.SetChatMetadataNumber(ctx, here, userID, "bad'key", 1); err == nil {
		t.Fatal("an unsafe metadata key was accepted")
	}
	for _, query := range []string{"espresso price", "spress"} {
		hits, err := repo.SearchChats(ctx, userID, query, 10)
		if err != nil {
			t.Fatalf("search %q: %v", query, err)
		}
		if len(hits) != 1 || hits[0].SessionID != here || hits[0].Title != "Pricing" || !strings.Contains(hits[0].Snippet, "4 dollars") {
			t.Fatalf("search %q = %#v", query, hits)
		}
	}
}

func TestMatchSnippetCentersOnTheMatch(t *testing.T) {
	text := strings.Repeat("intro words ", 60) + "the refund window is 30 days" + strings.Repeat(" trailing words", 60)
	snippet := matchSnippet(text, "refund", 120)
	if !strings.Contains(snippet, "refund window") || !strings.HasPrefix(snippet, "...") || !strings.HasSuffix(snippet, "...") {
		t.Fatalf("snippet = %q", snippet)
	}
}
