package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/security/internal/domain"
	"github.com/voxire/onion/internal/modules/security/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) GetSettings(ctx context.Context, userID int64) (domain.SecuritySettings, error) {
	return s.repo.GetSecuritySettings(ctx, userID)
}

func (s *Service) SetTwoFactor(ctx context.Context, userID int64, enabled bool) error {
	return s.repo.SetTwoFactor(ctx, userID, enabled)
}
