package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/support_inbox/internal/domain"
	"github.com/voxire/onion/internal/modules/support_inbox/internal/infrastructure"
)

type repository interface {
	ListInbox(ctx context.Context, businessID int64, page, pageSize int) ([]domain.InboxItem, int, error)
	UnreadCount(ctx context.Context, businessID int64) (int, error)
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) ListInbox(ctx context.Context, businessID int64, page, pageSize int) ([]domain.InboxItem, int, error) {
	return s.repo.ListInbox(ctx, businessID, page, pageSize)
}

func (s *Service) UnreadCount(ctx context.Context, businessID int64) (int, error) {
	return s.repo.UnreadCount(ctx, businessID)
}
