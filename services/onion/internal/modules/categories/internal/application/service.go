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

type repository interface {
	ListCategories(ctx context.Context, businessID int64) ([]domain.Category, error)
	ListCategoriesPaginated(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Category, error)
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
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListCategoriesPaginated(ctx, businessID, limit, offset)
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
