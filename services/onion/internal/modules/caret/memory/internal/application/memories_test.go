package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
)

type memoryRepository struct {
	repository
	savedID    int
	savedValue string
	deleted    int
}

func (r *memoryRepository) SaveMemory(_ context.Context, id int, value string) (*domain.UserMemory, error) {
	r.savedID, r.savedValue = id, value
	return &domain.UserMemory{ID: max(id, 1), Value: value}, nil
}

func (r *memoryRepository) DeleteMemory(_ context.Context, id int) error {
	r.deleted = id
	return nil
}

func TestSaveMemoryNormalizesAndValidates(t *testing.T) {
	repo := &memoryRepository{}
	service := NewService(repo)

	memory, err := service.SaveMemory(context.Background(), 0, "  Prefers   answers\nin French  ")
	if err != nil || memory.Value != "Prefers answers in French" || repo.savedID != 0 {
		t.Fatalf("memory=%+v err=%v savedID=%d", memory, err, repo.savedID)
	}

	for _, value := range []string{"", "   ", strings.Repeat("a", MaxMemoryRunes+1)} {
		if _, err := service.SaveMemory(context.Background(), 0, value); !errors.Is(err, ErrInvalidMemory) {
			t.Fatalf("value of %d chars: err=%v, want ErrInvalidMemory", len(value), err)
		}
	}
	if _, err := service.SaveMemory(context.Background(), -1, "ok"); !errors.Is(err, ErrInvalidMemory) {
		t.Fatal("negative id should be rejected")
	}
	if _, err := service.SaveMemory(context.Background(), 0, strings.Repeat("é", MaxMemoryRunes)); err != nil {
		t.Fatalf("limit counts characters, not bytes: %v", err)
	}
}

func TestDeleteMemoryRejectsInvalidIDs(t *testing.T) {
	repo := &memoryRepository{}
	service := NewService(repo)
	if err := service.DeleteMemory(context.Background(), 0); !errors.Is(err, ErrInvalidMemory) {
		t.Fatalf("err = %v", err)
	}
	if err := service.DeleteMemory(context.Background(), 12); err != nil || repo.deleted != 12 {
		t.Fatalf("err=%v deleted=%d", err, repo.deleted)
	}
}
