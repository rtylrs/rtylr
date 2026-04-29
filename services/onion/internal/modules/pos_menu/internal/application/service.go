package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos_menu/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_menu/internal/infrastructure"
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.MenuItem, error)
	QuickCreate(ctx context.Context, ownerID int64, input domain.MenuItemInput) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.MenuItem, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *Service) QuickCreate(ctx context.Context, ownerID int64, input domain.MenuItemInput) (int64, error) {
	return s.repo.QuickCreate(ctx, ownerID, input)
}
