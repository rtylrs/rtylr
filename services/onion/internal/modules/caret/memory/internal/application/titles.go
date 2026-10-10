package application

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
)

const (
	autoTitledKey     = "auto_titled"
	maxTitleRunes     = 60
	derivedTitleRunes = 80
	titleExcerptBytes = 1200
)

// TitleWork is the first exchange of a chat that still carries the title
// derived from its first message.
type TitleWork struct {
	CurrentTitle string
	FirstPrompt  string
	FirstReply   string
}

// derivedTitle matches how a new chat is titled from its first message.
func derivedTitle(prompt string) string {
	title := strings.Join(strings.Fields(prompt), " ")
	if title == "" {
		return "New chat"
	}
	if runes := []rune(title); len(runes) > derivedTitleRunes {
		return string(runes[:derivedTitleRunes]) + "..."
	}
	return title
}

func sessionMetadata(chat *domain.ChatSession) map[string]any {
	metadata := map[string]any{}
	if raw := strings.TrimSpace(chat.Metadata); raw != "" && raw != "null" {
		_ = json.Unmarshal([]byte(raw), &metadata)
	}
	return metadata
}

// PendingChatTitle reports whether a chat should get a written title: it has
// its first reply, was never auto-titled, and still has the title derived
// from its first message (so a chat the user renamed is left alone).
func (s *Service) PendingChatTitle(ctx context.Context, session string, userID int) (TitleWork, bool, error) {
	chat, err := s.repository.GetChatSession(ctx, session, userID)
	if err != nil {
		return TitleWork{}, false, err
	}
	if done, _ := sessionMetadata(chat)[autoTitledKey].(bool); done {
		return TitleWork{}, false, nil
	}
	messages, err := s.repository.GetContextMessages(ctx, session)
	if err != nil {
		return TitleWork{}, false, err
	}
	sort.SliceStable(messages, func(i, j int) bool { return messages[i].ID < messages[j].ID })
	var prompt, reply string
	for _, message := range messages {
		switch {
		case prompt == "" && strings.EqualFold(message.Role, "user"):
			prompt = message.Content
		case prompt != "" && reply == "" && strings.EqualFold(message.Role, "assistant"):
			reply = message.Content
		}
	}
	if prompt == "" || reply == "" {
		return TitleWork{}, false, nil
	}
	if current := strings.TrimSpace(chat.Title); current != "" && current != derivedTitle(prompt) && current != "New chat" {
		return TitleWork{}, false, nil
	}
	return TitleWork{
		CurrentTitle: chat.Title,
		FirstPrompt:  clipEnd(strings.TrimSpace(prompt), titleExcerptBytes),
		FirstReply:   clipEnd(strings.TrimSpace(reply), titleExcerptBytes),
	}, true, nil
}

// SaveChatTitle stores a written title, unless the chat was renamed after
// the title was requested. Either way the chat is marked as titled.
func (s *Service) SaveChatTitle(ctx context.Context, session string, userID int, expectedTitle, title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return errors.New("chat title is empty")
	}
	if runes := []rune(title); len(runes) > maxTitleRunes {
		title = strings.TrimSpace(string(runes[:maxTitleRunes]))
	}
	chat, err := s.repository.GetChatSession(ctx, session, userID)
	if err != nil {
		return err
	}
	metadata := sessionMetadata(chat)
	metadata[autoTitledKey] = true
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	if chat.Title != expectedTitle {
		title = chat.Title
	}
	defer invalidateMemoryScope(ctx, userID)
	_, err = s.repository.UpdateChatSession(ctx, session, userID, title, chat.Summary, string(encoded))
	return err
}
