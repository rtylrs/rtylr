package application

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/store_themes/internal/domain"
	"github.com/voxire/onion/internal/modules/store_themes/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const cacheTTL = 5 * time.Minute

type Service struct{ repo *infrastructure.Repository }

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Theme, error) {
	cacheKey := "store_themes:" + strconv.FormatInt(businessID, 10)
	return cacher.Of[[]domain.Theme]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.Theme, error) {
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
	cacher.Of[[]domain.Theme]().Invalidate(ctx, []string{"store_themes:" + strconv.FormatInt(businessID, 10)})
}
