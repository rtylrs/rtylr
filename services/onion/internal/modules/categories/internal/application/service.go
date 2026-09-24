package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/categories/internal/domain"
	"github.com/voxire/onion/internal/modules/categories/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/audit"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const (
	categoriesCacheTTL = 5 * time.Minute

	defaultListLimit int32 = 20
	maxListLimit     int32 = 100
)

type CategoryPage struct {
	Categories []domain.Category
	Total      int64
	Limit      int64
	Offset     int64
}

type repository interface {
	ListCategories(ctx context.Context, businessID int64) ([]domain.Category, error)
	ListCategoriesPaginated(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Category, error)
	CountCategories(ctx context.Context, businessID int64) (int64, error)
	CreateCategory(ctx context.Context, businessID int64, input domain.CategoryInput) (int64, error)
	UpdateCategory(ctx context.Context, businessID, id int64, input domain.CategoryInput) error
	DeleteCategory(ctx context.Context, businessID, id int64) error
}

type Service struct {
	repo repository
	db   *sql.DB
}

func NewService(_ context.Context, db *sql.DB) *Service {
	repo := infrastructure.NewRepository(db)

	return &Service{
		repo: repo,
		db:   db,
	}
}

func (s *Service) ListCategories(ctx context.Context, businessID int64) ([]domain.Category, error) {
	cacheKey := fmt.Sprintf("categories:%d", businessID)

	return cacher.Of[[]domain.Category]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.Category, error) {
		return s.repo.ListCategories(ctx, businessID)
	}, cacher.WithExpiry(ctx, categoriesCacheTTL))
}

// ListCategoriesPaginated returns a single page of categories. A non-positive
// limit defaults to 20 and limit is capped at 100; a negative offset is
// clamped to zero.
func (s *Service) ListCategoriesPaginated(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Category, error) {
	normalizedLimit, normalizedOffset := normalizeListBounds(int64(limit), int64(offset))
	return s.repo.ListCategoriesPaginated(ctx, businessID, normalizedLimit, normalizedOffset)
}

func (s *Service) ListCategoriesPage(ctx context.Context, businessID int64, limit, offset int64) (CategoryPage, error) {
	normalizedLimit, normalizedOffset := normalizeListBounds(limit, offset)
	categories, err := s.repo.ListCategoriesPaginated(ctx, businessID, normalizedLimit, normalizedOffset)
	if err != nil {
		return CategoryPage{}, err
	}
	total, err := s.repo.CountCategories(ctx, businessID)
	if err != nil {
		return CategoryPage{}, err
	}
	return CategoryPage{
		Categories: categories,
		Total:      total,
		Limit:      int64(normalizedLimit),
		Offset:     int64(normalizedOffset),
	}, nil
}

func normalizeListBounds(limit, offset int64) (int32, int32) {
	if limit <= 0 {
		limit = int64(defaultListLimit)
	}
	if limit > int64(maxListLimit) {
		limit = int64(maxListLimit)
	}
	if offset < 0 {
		offset = 0
	}
	const maxInt32 = int64(1<<31 - 1)
	if offset > maxInt32 {
		offset = maxInt32
	}
	return int32(limit), int32(offset)
}

func (s *Service) CreateCategory(ctx context.Context, businessID int64, input domain.CategoryInput) (int64, error) {
	id, err := s.repo.CreateCategory(ctx, businessID, input)
	if err != nil {
		return 0, err
	}
	if err := audit.Append(ctx, s.db, audit.Event{BusinessID: businessID, ActorID: businessID, EntityType: "category", EntityID: id, Action: "created", NewValues: input}); err != nil {
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
	if err := audit.Append(ctx, s.db, audit.Event{BusinessID: businessID, ActorID: businessID, EntityType: "category", EntityID: id, Action: "updated", NewValues: input}); err != nil {
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
	if err := audit.Append(ctx, s.db, audit.Event{BusinessID: businessID, ActorID: businessID, EntityType: "category", EntityID: id, Action: "deleted"}); err != nil {
		return err
	}

	s.invalidateCache(ctx, businessID)

	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	prefix := fmt.Sprintf("categories:%d", businessID)
	cacher.Of[[]domain.Category]().Invalidate(ctx, []string{prefix})
}
