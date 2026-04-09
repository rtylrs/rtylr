package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/careers/internal/domain"
	"github.com/voxire/onion/internal/modules/careers/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	repo := infrastructure.NewRepository(ctx)

	return &Service{
		repo: repo,
	}
}

func (s *Service) GetActiveCareers(ctx context.Context) ([]domain.Career, error) {
	return s.repo.GetActiveCareers(ctx)
}
