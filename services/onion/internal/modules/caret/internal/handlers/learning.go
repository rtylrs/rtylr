package handlers

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/caret/internal/application"
	"github.com/voxire/onion/internal/platform/logger"
)

// streamCompleter gives a streaming-only AI service the one-shot completion
// that background passes (lessons) need.
type streamCompleter struct{ ai AIStreamService }

func (c streamCompleter) CompleteText(ctx context.Context, prompt string) (string, error) {
	tokens, err := c.ai.GenerateTextStream(ctx, prompt)
	if err != nil {
		return "", err
	}
	var text strings.Builder
	for token := range tokens {
		if token.Error != "" {
			return "", errors.New(token.Error)
		}
		text.WriteString(token.Text)
	}
	return text.String(), nil
}

func completerFor(ai AIStreamService) application.TextCompleter {
	if completer, ok := ai.(application.TextCompleter); ok {
		return completer
	}
	return streamCompleter{ai: ai}
}

// learnFromCorrection records a message that pushes back on the previous
// reply as feedback on it, then learns from it in the background.
func (h *GenerateTextStreamHandler) learnFromCorrection(ctx context.Context, sessionID string, userID int, message string) {
	learner, ok := h.MemoryService.(application.LessonLearner)
	if !ok || h.AiService == nil || !h.AiService.IsAvailable() {
		return
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Ctx(ctx).Error("panic while learning from a correction", logger.Any("panic", r))
			}
		}()
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Minute)
		defer cancel()
		pending, err := learner.RecordReplyFeedback(ctx, sessionID, userID, "", "correction", message)
		if err != nil || !application.ShouldDistillLessons("correction", pending) {
			return
		}
		if err := application.DistillLessons(ctx, completerFor(h.AiService), learner, userID); err != nil {
			logger.Ctx(ctx).Warn("Caret could not learn from a correction", logger.Error(err))
		}
	}()
}

// learnFromChat reads the newest part of a chat for durable facts after a
// reply is saved; most passes find nothing and only record their progress.
func (h *GenerateTextStreamHandler) learnFromChat(ctx context.Context, learner application.ChatLearner, sessionID string, userID int) {
	defer func() {
		if r := recover(); r != nil {
			logger.Ctx(ctx).Error("panic while learning from a chat", logger.Any("panic", r))
		}
	}()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if err := application.LearnFromChat(ctx, completerFor(h.AiService), learner, sessionID, userID); err != nil {
		logger.Ctx(ctx).Warn("Caret could not learn from a chat", logger.Error(err))
	}
}
