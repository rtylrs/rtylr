package application

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
	"github.com/voxire/onion/internal/platform/auth"
)

// MaxMemoryRunes bounds one remembered note; memories are short facts and
// preferences, not documents.
const MaxMemoryRunes = 500

var ErrInvalidMemory = errors.New("memory must be 1 to 500 characters")

// ListMemories returns what Caret remembers for the signed-in user here.
func (s *Service) ListMemories(ctx context.Context) ([]domain.UserMemory, error) {
	return s.repository.ListMemories(ctx)
}

// SaveMemory edits a memory, or adds a personal one when id is 0.
func (s *Service) SaveMemory(ctx context.Context, id int, value string) (*domain.UserMemory, error) {
	value = strings.Join(strings.Fields(value), " ")
	if value == "" || utf8.RuneCountInString(value) > MaxMemoryRunes || id < 0 {
		return nil, ErrInvalidMemory
	}
	memory, err := s.repository.SaveMemory(ctx, id, value)
	if err == nil {
		invalidateCurrentUserMemory(ctx)
	}
	return memory, err
}

// DeleteMemory makes Caret forget one memory.
func (s *Service) DeleteMemory(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrInvalidMemory
	}
	err := s.repository.DeleteMemory(ctx, id)
	if err == nil {
		invalidateCurrentUserMemory(ctx)
	}
	return err
}

func invalidateCurrentUserMemory(ctx context.Context) {
	if user, ok := auth.GetAuthUser(ctx); ok && user != nil {
		invalidateMemoryScope(ctx, int(user.UserID))
	}
}
