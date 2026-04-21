package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/accounts/internal/domain"
	"github.com/voxire/onion/internal/modules/accounts/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Account, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.AccountInput) (int64, error) {
	return s.repo.Create(ctx, ownerID, input)
}

func (s *Service) Update(ctx context.Context, ownerID, id int64, input domain.AccountInput) error {
	return s.repo.Update(ctx, ownerID, id, input)
}

func (s *Service) Delete(ctx context.Context, ownerID, id int64) error {
	return s.repo.Delete(ctx, ownerID, id)
}
