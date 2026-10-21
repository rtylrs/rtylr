package application

import (
	"context"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/store/themes/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const cacheTTL = 5 * time.Minute

type Service struct{ repo repository }

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Theme, error) {
	cacheKey := "store_themes:" + strconv.FormatInt(businessID, 10)
	return cacher.Of[[]domain.Theme]().GetOrComputeSingle(ctx, cacheKey+"|domain=v2", func(ctx context.Context) ([]domain.Theme, error) {
		return s.repo.List(ctx, businessID)
	}, cacher.WithExpiry(ctx, cacheTTL))
}

func (s *Service) Create(ctx context.Context, creatorID, businessID int64, input domain.ThemeInput) (int64, error) {
	id, err := s.repo.Create(ctx, creatorID, businessID, input)
	if err != nil {
		return 0, err
	}
	s.invalidateCache(ctx, businessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.ThemeInput) error {
	if err := s.repo.Update(ctx, businessID, id, input); err != nil {
		return err
	}
	s.invalidateCache(ctx, businessID)
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	cacher.Of[[]domain.Theme]().Invalidate(ctx, []string{"store_themes:" + strconv.FormatInt(businessID, 10) + "|domain=v2"})
}

type repository interface {
	Create(ctx context.Context, creatorID, businessID int64, input domain.ThemeInput) (int64, error)
	List(ctx context.Context, businessID int64) ([]domain.Theme, error)
	Update(ctx context.Context, businessID, id int64, input domain.ThemeInput) error
}
