package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/categories/internal/domain"
	"github.com/voxire/onion/internal/modules/categories/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const categoriesCacheTTL = 5 * time.Minute

type repository interface {
	ListCategories(ctx context.Context, ownerID int64) ([]domain.Category, error)
	CreateCategory(ctx context.Context, ownerID int64, input domain.CategoryInput) (int64, error)
	UpdateCategory(ctx context.Context, ownerID, id int64, input domain.CategoryInput) error
	DeleteCategory(ctx context.Context, ownerID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	repo := infrastructure.NewRepository(ctx)

	return &Service{
		repo: repo,
	}
}

func (s *Service) ListCategories(ctx context.Context, ownerID int64) ([]domain.Category, error) {
	cacheKey := fmt.Sprintf("categories:%d", ownerID)

	return cacher.Of[[]domain.Category]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.Category, error) {
		return s.repo.ListCategories(ctx, ownerID)
	}, cacher.WithExpiry(ctx, categoriesCacheTTL))
}

func (s *Service) CreateCategory(ctx context.Context, ownerID int64, input domain.CategoryInput) (int64, error) {
	id, err := s.repo.CreateCategory(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}

	s.invalidateCache(ctx, ownerID)

	return id, nil
}

func (s *Service) UpdateCategory(ctx context.Context, ownerID, id int64, input domain.CategoryInput) error {
	err := s.repo.UpdateCategory(ctx, ownerID, id, input)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, ownerID)

	return nil
}

func (s *Service) DeleteCategory(ctx context.Context, ownerID, id int64) error {
	err := s.repo.DeleteCategory(ctx, ownerID, id)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, ownerID)

	return nil
}

func (s *Service) invalidateCache(ctx context.Context, ownerID int64) {
	prefix := fmt.Sprintf("categories:%d", ownerID)
	cacher.Of[[]domain.Category]().Invalidate(ctx, []string{prefix})
}
