package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos_floors/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_floors/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, businessID, branchID int64) ([]domain.Floor, error) {
	return s.repo.List(ctx, businessID, branchID)
}

func (s *Service) Create(ctx context.Context, input domain.FloorInput) (int64, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, input domain.FloorUpdateInput) error {
	return s.repo.Update(ctx, input)
}

func (s *Service) Delete(ctx context.Context, businessID, id int64) error {
	return s.repo.Delete(ctx, businessID, id)
}
