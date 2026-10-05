package infrastructure

import (
	"context"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/voxire/onion/internal/modules/organization/scopes"
	"github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/testutil"
)

func TestAssistantMemoriesScopesFactsAndLoadsPreferences(t *testing.T) {
	if host := os.Getenv("DB_HOST"); host != "" && host != "localhost" && host != "127.0.0.1" {
		t.Fatal("memory fixtures require a local test database")
	}
	if name := os.Getenv("DB_NAME"); name != "" && name != "onion_test" {
		t.Fatal("memory fixtures require onion_test")
	}
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const userID, otherUserID, businessID, otherBusinessID = 741231, 741232, 910071, 910072
	suffix := uuid.NewString()
	if _, err := db.ExecContext(ctx, `INSERT INTO users (id,business_id,email,password_hash) VALUES (?, ?, ?, 'fixture'), (?, ?, ?, 'fixture')`,
		userID, businessID, suffix+"@caret.test", otherUserID, businessID, "other-"+suffix+"@caret.test"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, query := range []string{
			"DELETE FROM user_memory WHERE user_id IN (?,?)",
			"DELETE FROM chat_sessions WHERE user_id IN (?,?)",
			"DELETE FROM users WHERE id IN (?,?)",
		} {
			if _, err := db.ExecContext(ctx, query, userID, otherUserID); err != nil {
				t.Errorf("remove memory fixtures: %v", err)
			}
		}
	})
	current, same, foreign, legacy := "current-"+suffix, "same-"+suffix, "foreign-"+suffix, "legacy-"+suffix
	if _, err := db.ExecContext(ctx, `INSERT INTO chat_sessions (user_id,session_id,metadata) VALUES
		(?,?,JSON_OBJECT('business_id',?)),(?,?,JSON_OBJECT('business_id',?)),
		(?,?,JSON_OBJECT('business_id',?)),(?,?,JSON_OBJECT())`,
		userID, current, businessID, userID, same, businessID, userID, foreign, otherBusinessID, userID, legacy); err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		user       int
		kind, key  string
		source     any
		confidence float64
	}{
		{userID, "preference", "global", nil, 1},
		{userID, "fact", "same-business", same, 1},
		{userID, "fact", "foreign-business", foreign, 1},
		{userID, "fact", "legacy-business", legacy, 1},
		{userID, "fact", "low-confidence", same, 0.6},
		{userID, "fact", "missing-source", nil, 1},
		{otherUserID, "preference", "other-user", nil, 1},
	} {
		if _, err := db.ExecContext(ctx, "INSERT INTO user_memory (user_id,memory_type,`key`,value,source_session,confidence) VALUES (?,?,?,'fixture',?,?)", row.user, row.kind, row.key, row.source, row.confidence); err != nil {
			t.Fatal(err)
		}
	}
	repo := NewRepository(db).WithIdentityReaders(scopes.NewReader(db).DefaultBusiness, scopes.NewReader(db).AssistantIdentity)
	for _, tc := range []struct {
		name, session string
		business      int64
		want          []string
	}{
		{"current business", current, businessID, []string{"global", "legacy-business", "same-business"}},
		{"legacy chat", legacy, businessID, []string{"global", "legacy-business", "same-business"}},
		{"second business", foreign, otherBusinessID, []string{"foreign-business", "global"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := auth.WithAuthUser(ctx, &auth.AuthUser{UserID: userID, BusinessID: tc.business})
			memories, err := repo.AssistantMemories(ctx, tc.session)
			if err != nil {
				t.Fatalf("load assistant memories: %v", err)
			}
			keys := make([]string, 0, len(memories))
			for _, memory := range memories {
				keys = append(keys, memory.Key)
			}
			sort.Strings(keys)
			if !reflect.DeepEqual(keys, tc.want) {
				t.Fatalf("memories = %v, want %v", keys, tc.want)
			}
		})
	}
}
