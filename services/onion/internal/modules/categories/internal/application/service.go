package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/categories/internal/domain"
	"github.com/voxire/onion/internal/modules/categories/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const categoriesCacheTTL = 5 * time.Minute

type repository interface {
	ListCategories(ctx context.Context, businessID int64) ([]domain.Category, error)
	CreateCategory(ctx context.Context, businessID int64, input domain.CategoryInput) (int64, error)
	UpdateCategory(ctx context.Context, businessID, id int64, input domain.CategoryInput) error
	DeleteCategory(ctx context.Context, businessID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	repo := infrastructure.NewRepository(db)

	return &Service{
		repo: repo,
	}
}

func (s *Service) ListCategories(ctx context.Context, businessID int64) ([]domain.Category, error) {
	cacheKey := fmt.Sprintf("categories:%d", businessID)

	return cacher.Of[[]domain.Category]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.Category, error) {
		return s.repo.ListCategories(ctx, businessID)
	}, cacher.WithExpiry(ctx, categoriesCacheTTL))
}

func (s *Service) CreateCategory(ctx context.Context, businessID int64, input domain.CategoryInput) (int64, error) {
	id, err := s.repo.CreateCategory(ctx, businessID, input)
	if err != nil {
		return 0, err
	}

	s.invalidateCache(ctx, businessID)

	return id, nil
}

func (s *Service) UpdateCategory(ctx context.Context, businessID, id int64, input domain.CategoryInput) error {
	err := s.repo.UpdateCategory(ctx, businessID, id, input)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, businessID)

	return nil
}

func (s *Service) DeleteCategory(ctx context.Context, businessID, id int64) error {
	err := s.repo.DeleteCategory(ctx, businessID, id)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, businessID)

	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	prefix := fmt.Sprintf("categories:%d", businessID)
	cacher.Of[[]domain.Category]().Invalidate(ctx, []string{prefix})
}
