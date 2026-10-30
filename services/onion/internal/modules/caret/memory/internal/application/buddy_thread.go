package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
)

// BuddyThreadPrefix marks the one chat each person has with their buddy. The
// regular chat list leaves these out because the buddy has its own entry.
const BuddyThreadPrefix = "buddy-"

func BuddyThreadID(businessID int64, userID int) string {
	return fmt.Sprintf("%s%d-%d", BuddyThreadPrefix, businessID, userID)
}

// EnsureBuddyThread returns the person's buddy chat, creating it on first use.
func (s *Service) EnsureBuddyThread(ctx context.Context, businessID int64, userID int, title string) (string, error) {
	sessionID := BuddyThreadID(businessID, userID)
	if _, err := s.repository.GetChatSession(ctx, sessionID, userID); err == nil {
		return sessionID, nil
	} else if !errors.Is(err, domain.ErrChatSessionNotFound) {
		return "", err
	}
	metadata, _ := json.Marshal(map[string]any{"business_id": businessID, "buddy": true})
	now := time.Now()
	_, err := s.repository.CreateChatSession(ctx, &domain.ChatSession{
		UserID: userID, SessionID: sessionID, Title: title, Metadata: string(metadata), CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		// Another request may have created it first.
		if _, getErr := s.repository.GetChatSession(ctx, sessionID, userID); getErr == nil {
			return sessionID, nil
		}
		return "", err
	}
	invalidateMemoryScope(ctx, userID)
	return sessionID, nil
}

// RecentMessages returns up to limit of the newest messages in a chat the
// person owns, oldest first.
func (s *Service) RecentMessages(ctx context.Context, sessionID string, userID, limit int) ([]domain.ConversationMessage, error) {
	if _, err := s.repository.GetChatSession(ctx, sessionID, userID); err != nil {
		return nil, err
	}
	messages, err := s.repository.GetSessionMessages(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if len(messages) > limit {
		messages = messages[len(messages)-limit:]
	}
	return messages, nil
}

// BuddyLastPost is when the buddy last wrote in a chat the person owns.
func (s *Service) BuddyLastPost(ctx context.Context, sessionID string, userID int) (time.Time, bool, error) {
	if _, err := s.repository.GetChatSession(ctx, sessionID, userID); err != nil {
		return time.Time{}, false, err
	}
	return s.repository.LatestAssistantAt(ctx, sessionID)
}
