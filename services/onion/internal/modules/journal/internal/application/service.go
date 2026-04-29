package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/journal/internal/domain"
	"github.com/voxire/onion/internal/modules/journal/internal/infrastructure"
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.JournalEntry, error)
	Create(ctx context.Context, ownerID int64, input domain.JournalEntryInput) (int64, error)
	Post(ctx context.Context, ownerID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.JournalEntry, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.JournalEntryInput) (int64, error) {
	return s.repo.Create(ctx, ownerID, input)
}

func (s *Service) Post(ctx context.Context, ownerID, id int64) error {
	return s.repo.Post(ctx, ownerID, id)
}
