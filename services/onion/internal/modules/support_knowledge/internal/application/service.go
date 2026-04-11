package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/support_knowledge/internal/domain"
	"github.com/voxire/onion/internal/modules/support_knowledge/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) List(ctx context.Context, businessID int64, page, pageSize int) ([]domain.Article, int, error) {
	return s.repo.List(ctx, businessID, page, pageSize)
}

func (s *Service) Create(ctx context.Context, input domain.ArticleInput) (int64, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.ArticleInput) error {
	return s.repo.Update(ctx, businessID, id, input)
}

func (s *Service) Delete(ctx context.Context, businessID, id int64) error {
	return s.repo.Delete(ctx, businessID, id)
}
