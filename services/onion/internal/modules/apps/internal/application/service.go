package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/apps/internal/domain"
	"github.com/voxire/onion/internal/modules/apps/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) ListApps() []domain.App {
	return domain.AvailableApps()
}
