package application

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
)

const (
	maxFeedbackReasonRunes = 1000
	feedbackExcerptBytes   = 2500
	// MaxLessons bounds how many lessons Caret carries into every prompt.
	MaxLessons           = 15
	maxLessonRunes       = 300
	pendingFeedbackBatch = 20
)

var ErrInvalidFeedback = errors.New("rating must be up, down, clear or correction, with a reason under 1,000 characters")

// RecordReplyFeedback stores the person's verdict on a reply in one of their
// chats and reports how many verdicts are waiting to be learned from. The
// reply is read from the chat, never taken from the client: messageID is a
// stored message id, or empty or unknown for the chat's latest reply.
func (s *Service) RecordReplyFeedback(ctx context.Context, sessionID string, userID int, messageID, rating, reason string) (int, error) {
	rating = strings.ToLower(strings.TrimSpace(rating))
	reason = strings.Join(strings.Fields(reason), " ")
	switch rating {
	case "up", "down", "clear", "correction":
	default:
		return 0, ErrInvalidFeedback
	}
	if utf8.RuneCountInString(reason) > maxFeedbackReasonRunes || (rating == "correction" && reason == "") {
		return 0, ErrInvalidFeedback
	}
	chat, err := s.repository.GetChatSession(ctx, strings.TrimSpace(sessionID), userID)
	if err != nil {
		return 0, err
	}
	messages, err := s.repository.GetContextMessages(ctx, chat.SessionID)
	if err != nil {
		return 0, err
	}
	prompt, reply, ok := ratedExchange(messages, messageID)
	if !ok {
		return 0, ErrChatSessionNotFound
	}
	feedback := domain.ReplyFeedback{
		SessionID: chat.SessionID, MessageID: int64(reply.ID), Rating: rating, Reason: reason,
		Prompt: clipEnd(prompt.Content, feedbackExcerptBytes), Reply: clipMiddle(reply.Content, feedbackExcerptBytes),
	}
	if err := s.repository.SaveReplyFeedback(ctx, userID, feedback); err != nil {
		return 0, err
	}
	pending, err := s.repository.PendingReplyFeedback(ctx, userID, pendingFeedbackBatch)
	return len(pending), err
}

// ratedExchange finds the rated reply (by stored id, else the latest reply)
// and the user message it answered.
func ratedExchange(messages []domain.ConversationMessage, messageID string) (prompt, reply domain.ConversationMessage, ok bool) {
	ordered := append([]domain.ConversationMessage(nil), messages...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	target := -1
	if id, err := strconv.Atoi(strings.TrimSpace(messageID)); err == nil {
		for i, message := range ordered {
			if message.ID == id && strings.EqualFold(message.Role, "assistant") {
				target = i
			}
		}
	}
	if target < 0 {
		for i := len(ordered) - 1; i >= 0; i-- {
			if strings.EqualFold(ordered[i].Role, "assistant") {
				target = i
				break
			}
		}
	}
	if target < 0 {
		return prompt, reply, false
	}
	for i := target - 1; i >= 0; i-- {
		if strings.EqualFold(ordered[i].Role, "user") {
			prompt = ordered[i]
			break
		}
	}
	return prompt, ordered[target], true
}

// LessonWork is what a distillation pass reads: the current lessons and the
// verdicts not yet learned from.
type LessonWork struct {
	Lessons  []string
	Feedback []domain.ReplyFeedback
}

// PendingLessonWork returns the lessons and the pending feedback for a user.
func (s *Service) PendingLessonWork(ctx context.Context, userID int) (LessonWork, error) {
	feedback, err := s.repository.PendingReplyFeedback(ctx, userID, pendingFeedbackBatch)
	if err != nil || len(feedback) == 0 {
		return LessonWork{}, err
	}
	lessons, err := s.repository.UserLessons(ctx, userID)
	if err != nil {
		return LessonWork{}, err
	}
	return LessonWork{Lessons: lessons, Feedback: feedback}, nil
}

// SaveLessons replaces the user's lessons with a distilled set and marks the
// feedback they came from as learned.
func (s *Service) SaveLessons(ctx context.Context, userID int, lessons []string, feedbackIDs []int64) error {
	clean := make([]string, 0, len(lessons))
	seen := map[string]bool{}
	for _, lesson := range lessons {
		lesson = strings.Join(strings.Fields(lesson), " ")
		if runes := []rune(lesson); len(runes) > maxLessonRunes {
			lesson = string(runes[:maxLessonRunes])
		}
		key := strings.ToLower(lesson)
		if lesson == "" || seen[key] {
			continue
		}
		seen[key] = true
		clean = append(clean, lesson)
		if len(clean) == MaxLessons {
			break
		}
	}
	if err := s.repository.ReplaceLessons(ctx, userID, clean, feedbackIDs); err != nil {
		return err
	}
	invalidateMemoryScope(ctx, userID)
	return nil
}
