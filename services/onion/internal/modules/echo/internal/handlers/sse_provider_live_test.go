package handlers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/memory"
	"github.com/voxire/onion/internal/platform/snips"
	"github.com/voxire/onion/internal/testutil"
)

func TestLiveProviderPersistence(t *testing.T) {
	if os.Getenv("CARET_LIVE_PROVIDER_TEST") != "1" {
		t.Skip("set CARET_LIVE_PROVIDER_TEST=1 with provider keys and a local fixture database")
	}
	if host := os.Getenv("DB_HOST"); (host != "127.0.0.1" && host != "localhost") || os.Getenv("DB_NAME") != "onion_test" {
		t.Fatal("live provider verification requires the local onion_test fixture database")
	}
	db := testutil.OpenDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	const userID = 741205
	testutil.GrantBusinessAccess(t, db, userID, userID)
	m := memory.NewModule(ctx, db)
	session, err := m.CreateSession(ctx, userID, "Synthetic provider persistence check")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := m.GetService().DeleteSession(context.Background(), session, userID); err != nil {
			t.Errorf("remove fixture session: %v", err)
		}
	})
	ctx = snips.WithConversation(ctx, session)
	prompt := "Write one short sentence explaining what a project task is."
	if _, err := m.StoreMessage(ctx, session, userID, "user", prompt, 16); err != nil {
		t.Fatal(err)
	}
	client, err := snips.NewClient(ctx)
	if err != nil {
		t.Fatal("provider initialization failed")
	}
	t.Cleanup(func() { _ = client.Close() })
	planned, err := client.CompleteTextWithOptions(ctx, `Return a JSON object with a single key "ready" set to true.`, snips.DeterministicJSONCompletionOptions())
	if err != nil {
		t.Fatal("provider JSON completion failed")
	}
	var plan struct {
		Ready bool `json:"ready"`
	}
	if json.Unmarshal([]byte(planned), &plan) != nil || !plan.Ready {
		t.Fatal("provider did not return the requested JSON object")
	}
	tokens, err := client.CompleteTextStream(ctx, prompt)
	if err != nil {
		t.Fatal("provider stream initialization failed")
	}
	rec := httptest.NewRecorder()
	h := &GenerateTextStreamHandler{MemoryService: m}
	h.streamTokensToClient(ctx, rec, tokens, make(chan bool), &GenerateTextStreamRequest{SessionID: session, UserID: userID}, userID, userID, cacheWritePlan{}, nil)
	if strings.Contains(rec.Body.String(), `"type":"error"`) || !strings.Contains(rec.Body.String(), `"type":"done"`) {
		t.Fatal("provider stream did not complete successfully")
	}
	var content string
	if err := db.QueryRowContext(ctx, "SELECT content FROM conversation_messages WHERE session_id=? AND user_id=? AND role='assistant'", session, userID).Scan(&content); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(content) == "" {
		t.Fatal("stored answer is empty")
	}
	t.Logf("Provider stream completed and assistant answer reloaded (%d characters)", len(content))
}
