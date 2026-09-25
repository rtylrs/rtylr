package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/reviews/internal/domain"
	"github.com/voxire/onion/internal/modules/reviews/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	reviewsScope = "reviews"
	reviewsTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.Review, error)
	Create(ctx context.Context, businessID int64, input domain.ReviewInput) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateReviewsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(reviewsScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Review, error) {
	return cacher.Of[[]domain.Review]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(reviewsScope, businessID, "list"),
		func(ctx context.Context) ([]domain.Review, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, reviewsTTL),
	)
}

func (s *Service) Create(ctx context.Context, businessID int64, input domain.ReviewInput) (int64, error) {
	id, err := s.repo.Create(ctx, businessID, input)
	if err != nil {
		return 0, err
	}
	invalidateReviewsScope(ctx, businessID)
	return id, nil
}
