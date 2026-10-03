package handlers

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/memory"
	"github.com/voxire/onion/internal/platform/aicache"
	"github.com/voxire/onion/internal/testutil"
)

func TestCaretReplyPersistsInChatHistory(t *testing.T) {
	for _, cached := range []bool{false, true} {
		t.Run(map[bool]string{false: "direct", true: "cached"}[cached], func(t *testing.T) {
			db := testutil.OpenDB(t)
			ctx := context.Background()
			const userID = 741204
			testutil.GrantBusinessAccess(t, db, userID, userID)
			memoryService := memory.NewModule(ctx, db)
			sessionID, err := memoryService.CreateSession(ctx, userID, "Caret persistence fixture")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := memoryService.GetService().DeleteSession(ctx, sessionID, userID); err != nil {
					t.Errorf("delete fixture session: %v", err)
				}
			})
			if _, err := memoryService.StoreMessage(ctx, sessionID, userID, "user", "test", 1); err != nil {
				t.Fatal(err)
			}

			rec := httptest.NewRecorder()
			handler := &GenerateTextStreamHandler{MemoryService: memoryService}
			if cached {
				handler.streamCachedResponse(ctx, rec, make(chan bool), &GenerateTextStreamRequest{SessionID: sessionID, UserID: userID}, aicache.ExactResponse{Response: caretSmokeCheckReply, Provider: "fixture"}, cacheWritePlan{}, userID, userID)
			} else {
				handler.streamDirectResponse(ctx, rec, &GenerateTextStreamRequest{SessionID: sessionID, UserID: userID}, caretSmokeCheckReply)
			}
			if strings.Contains(rec.Body.String(), `"type":"error"`) || !strings.Contains(rec.Body.String(), `"type":"done"`) {
				t.Fatalf("reply did not complete successfully: %s", rec.Body.String())
			}

			var content string
			err = db.QueryRowContext(ctx, "SELECT content FROM conversation_messages WHERE session_id = ? AND user_id = ? AND role = 'assistant'", sessionID, userID).Scan(&content)
			if err != nil {
				t.Fatalf("read persisted assistant reply: %v", err)
			}
			if content != caretSmokeCheckReply {
				t.Fatalf("persisted reply = %q, want %q", content, caretSmokeCheckReply)
			}
		})
	}
}
