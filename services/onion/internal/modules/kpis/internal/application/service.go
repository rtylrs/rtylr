package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/kpis/internal/domain"
	"github.com/voxire/onion/internal/modules/kpis/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) GetKPIs(ctx context.Context, ownerID int64) (domain.KPISummary, error) {
	return s.repo.GetKPIs(ctx, ownerID)
}
