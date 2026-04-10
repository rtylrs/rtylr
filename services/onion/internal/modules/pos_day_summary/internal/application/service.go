package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos_day_summary/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_day_summary/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) GetDaySummary(ctx context.Context, ownerID int64, date string) (domain.DaySummary, error) {
	return s.repo.GetDaySummary(ctx, ownerID, date)
}
