package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos_tables/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_tables/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Table, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *Service) Create(ctx context.Context, input domain.TableInput) (int64, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, input domain.TableUpdateInput) error {
	return s.repo.Update(ctx, input)
}

func (s *Service) Delete(ctx context.Context, ownerID, id int64) error {
	return s.repo.Delete(ctx, ownerID, id)
}
