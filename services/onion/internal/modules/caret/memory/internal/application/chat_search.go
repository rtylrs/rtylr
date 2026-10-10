package application

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
)

var ErrInvalidChatSearch = errors.New("search past chats with 2 to 200 characters")

// SearchChats finds what the user discussed before, in this business's chats.
func (s *Service) SearchChats(ctx context.Context, userID int, query string, limit int) ([]domain.ChatSearchHit, error) {
	query = strings.Join(strings.Fields(query), " ")
	if n := utf8.RuneCountInString(query); n < 2 || n > 200 || userID <= 0 {
		return nil, ErrInvalidChatSearch
	}
	if limit <= 0 || limit > 20 {
		limit = 10
	}
	return s.repository.SearchChats(ctx, userID, query, limit)
}

// ReadChat returns the latest messages of one of the user's chats here.
func (s *Service) ReadChat(ctx context.Context, sessionID string, userID, limit int) (*domain.ChatSession, []domain.ConversationMessage, error) {
	session, err := s.repository.GetChatSession(ctx, strings.TrimSpace(sessionID), userID)
	if err != nil {
		return nil, nil, err
	}
	messages, err := s.repository.GetSessionMessages(ctx, session.SessionID)
	if err != nil {
		return nil, nil, err
	}
	if limit <= 0 || limit > 40 {
		limit = 20
	}
	if len(messages) > limit {
		messages = messages[len(messages)-limit:]
	}
	return session, messages, nil
}
