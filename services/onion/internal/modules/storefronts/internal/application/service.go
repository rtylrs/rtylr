package application

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/storefronts/internal/domain"
	"github.com/voxire/onion/internal/modules/storefronts/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const cacheTTL = 5 * time.Minute

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Storefront, error) {
	cacheKey := "storefronts:" + strconv.FormatInt(businessID, 10)
	return cacher.Of[[]domain.Storefront]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.Storefront, error) {
		return s.repo.List(ctx, businessID)
	}, cacher.WithExpiry(ctx, cacheTTL))
}

func (s *Service) Create(ctx context.Context, creatorID, businessID int64, input domain.StorefrontInput) (int64, error) {
	if err := s.verifyThemeOwnership(ctx, businessID, input.ThemeID); err != nil {
		return 0, err
	}
	id, err := s.repo.Create(ctx, creatorID, businessID, input)
	if err != nil {
		return 0, err
	}
	s.invalidateCache(ctx, businessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.StorefrontInput) error {
	if err := s.verifyThemeOwnership(ctx, businessID, input.ThemeID); err != nil {
		return err
	}
	err := s.repo.Update(ctx, businessID, id, input)
	if err != nil {
		return err
	}
	s.invalidateCache(ctx, businessID)
	return nil
}

// verifyThemeOwnership ensures a referenced theme belongs to the same business.
// A zero themeID means no theme is linked, which is allowed.
func (s *Service) verifyThemeOwnership(ctx context.Context, businessID, themeID int64) error {
	if themeID == 0 {
		return nil
	}
	ok, err := s.repo.ThemeBelongsToOwner(ctx, businessID, themeID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrThemeNotFound
	}
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	prefix := "storefronts:" + strconv.FormatInt(businessID, 10)
	cacher.Of[[]domain.Storefront]().Invalidate(ctx, []string{prefix})
}
