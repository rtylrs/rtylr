package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/crm_settings/internal/domain"
	"github.com/voxire/onion/internal/modules/crm_settings/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) Get(ctx context.Context, ownerID int64) (domain.Settings, error) {
	return s.repo.Get(ctx, ownerID)
}

func (s *Service) Upsert(ctx context.Context, input domain.SettingsInput) error {
	return s.repo.Upsert(ctx, input)
}
