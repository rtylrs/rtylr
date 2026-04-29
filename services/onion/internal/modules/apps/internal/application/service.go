package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/apps/internal/domain"
	"github.com/voxire/onion/internal/modules/apps/internal/infrastructure"
)

type repository interface {
	GetEnabledAddons(ctx context.Context, businessID int64) map[string]bool
	GetUserAllowedApps(ctx context.Context, businessID, userID int64) []string
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) ListApps() []domain.App {
	return domain.AvailableApps()
}

func (s *Service) GetEnabledAddons(ctx context.Context, businessID int64) map[string]bool {
	return s.repo.GetEnabledAddons(ctx, businessID)
}

func (s *Service) GetUserAllowedApps(ctx context.Context, businessID, userID int64) []string {
	return s.repo.GetUserAllowedApps(ctx, businessID, userID)
}
