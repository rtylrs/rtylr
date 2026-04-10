package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos_manager/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_manager/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) VerifyPIN(ctx context.Context, ownerID int64, pin string) (*domain.VerifyResult, error) {
	return s.repo.VerifyPIN(ctx, ownerID, pin)
}
