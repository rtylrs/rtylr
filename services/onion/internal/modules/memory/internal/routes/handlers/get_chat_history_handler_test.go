package handlers

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/memory/internal/domain"
)

func TestToChatSummariesMapsDashShape(t *testing.T) {
	updated := time.Date(2026, 6, 10, 9, 30, 0, 0, time.UTC)
	sessions := []domain.ChatSession{
		{
			SessionID: "session-1",
			Title:     "Launch notes",
			Summary:   "Discussed launch work.",
			UpdatedAt: updated,
		},
		{
			SessionID: "session-2",
			UpdatedAt: updated.Add(time.Minute),
		},
	}

	chats := toChatSummaries(sessions)
	if len(chats) != 2 {
		t.Fatalf("expected 2 chats, got %d", len(chats))
	}
	if chats[0].ID != "session-1" || chats[0].Title != "Launch notes" || chats[0].Summary != "Discussed launch work." {
		t.Fatalf("first chat was not mapped correctly: %#v", chats[0])
	}
	if chats[1].Title != "Untitled chat" {
		t.Fatalf("empty titles should fall back, got %q", chats[1].Title)
	}
}

func TestToMessagesMapsDashShape(t *testing.T) {
	created := time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC)
	messages := []domain.ConversationMessage{
		{
			ID:         42,
			Role:       "assistant",
			Content:    "Done.",
			TokensUsed: 12,
			CreatedAt:  created,
		},
	}

	got := toMessages(messages)
	if len(got) != 1 {
		t.Fatalf("expected 1 message, got %d", len(got))
	}
	if got[0].ID != "42" || got[0].Role != "assistant" || got[0].Content != "Done." || got[0].TokensUsed != 12 {
		t.Fatalf("message was not mapped correctly: %#v", got[0])
	}
	if !got[0].CreatedAt.Equal(created) {
		t.Fatalf("createdAt mismatch: %s", got[0].CreatedAt)
	}
}
