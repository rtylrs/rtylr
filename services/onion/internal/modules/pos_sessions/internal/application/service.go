package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos_sessions/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_sessions/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Session, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *Service) Open(ctx context.Context, input domain.OpenSessionInput) (int64, error) {
	return s.repo.Open(ctx, input)
}

func (s *Service) Close(ctx context.Context, input domain.CloseSessionInput) error {
	return s.repo.Close(ctx, input)
}
