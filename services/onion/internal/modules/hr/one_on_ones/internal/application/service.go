package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/hr/one_on_ones/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	oneOnOnesScope = "one_on_ones"
	oneOnOnesTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.OneOnOne, error)
	Create(ctx context.Context, businessID int64, input domain.OneOnOneInput) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func invalidateOneOnOnesScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(oneOnOnesScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.OneOnOne, error) {
	return cacher.Of[[]domain.OneOnOne]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(oneOnOnesScope, businessID, "list")+"|domain=v2",
		func(ctx context.Context) ([]domain.OneOnOne, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, oneOnOnesTTL),
	)
}

func (s *Service) Create(ctx context.Context, businessID int64, input domain.OneOnOneInput) (int64, error) {
	id, err := s.repo.Create(ctx, businessID, input)
	if err != nil {
		return 0, err
	}
	invalidateOneOnOnesScope(ctx, businessID)
	return id, nil
}
