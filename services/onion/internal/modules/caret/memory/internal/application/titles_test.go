package application

import (
	"context"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
)

func titledChat(title string) *chatRepository {
	return &chatRepository{
		chat: domain.ChatSession{SessionID: "s", Title: title, Metadata: "{}"},
		messages: []domain.ConversationMessage{
			{ID: 1, Role: "user", Content: "hey can you help me plan the launch on friday"},
			{ID: 2, Role: "assistant", Content: "Sure, here is a plan."},
		},
	}
}

func TestChatTitleReplacesTheDerivedTitleOnce(t *testing.T) {
	repo := titledChat("hey can you help me plan the launch on friday")
	service := NewService(repo)

	work, ok, err := service.PendingChatTitle(context.Background(), "s", 1)
	if err != nil || !ok || work.FirstReply != "Sure, here is a plan." {
		t.Fatalf("work=%+v ok=%v err=%v", work, ok, err)
	}
	if err := service.SaveChatTitle(context.Background(), "s", 1, work.CurrentTitle, "Friday launch plan"); err != nil {
		t.Fatal(err)
	}
	if repo.chat.Title != "Friday launch plan" || !strings.Contains(repo.chat.Metadata, `"auto_titled":true`) {
		t.Fatalf("title not saved: %q %s", repo.chat.Title, repo.chat.Metadata)
	}
	if _, ok, _ := service.PendingChatTitle(context.Background(), "s", 1); ok {
		t.Fatal("a chat is titled only once")
	}
}

func TestChatTitleLeavesRenamedChatsAlone(t *testing.T) {
	repo := titledChat("Launch notes")
	if _, ok, _ := NewService(repo).PendingChatTitle(context.Background(), "s", 1); ok {
		t.Fatal("a chat the user renamed must keep its name")
	}

	// Renamed while the title was being written: the rename wins.
	repo = titledChat("hey can you help me plan the launch on friday")
	service := NewService(repo)
	work, _, _ := service.PendingChatTitle(context.Background(), "s", 1)
	repo.chat.Title = "My launch"
	_ = service.SaveChatTitle(context.Background(), "s", 1, work.CurrentTitle, "Friday launch plan")
	if repo.chat.Title != "My launch" {
		t.Fatalf("rename overwritten: %q", repo.chat.Title)
	}
}

func TestChatTitleWaitsForTheFirstReply(t *testing.T) {
	repo := titledChat("hey")
	repo.messages = repo.messages[:1]
	if _, ok, _ := NewService(repo).PendingChatTitle(context.Background(), "s", 1); ok {
		t.Fatal("no title before the first reply")
	}
}
