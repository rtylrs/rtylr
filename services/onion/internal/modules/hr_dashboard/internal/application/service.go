package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/hr_dashboard/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_dashboard/internal/infrastructure"
)

type repository interface {
	GetWidgets(ctx context.Context, ownerID int64) ([]*domain.DashboardWidget, error)
	UpsertWidgets(ctx context.Context, ownerID int64, widgets []*domain.DashboardWidget) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) GetWidgets(ctx context.Context, ownerID int64) ([]*domain.DashboardWidget, error) {
	return s.repo.GetWidgets(ctx, ownerID)
}

func (s *Service) UpsertWidgets(ctx context.Context, ownerID int64, widgets []*domain.DashboardWidget) error {
	return s.repo.UpsertWidgets(ctx, ownerID, widgets)
}
