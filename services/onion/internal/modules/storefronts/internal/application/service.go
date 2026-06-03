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

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Storefront, error) {
	cacheKey := "storefronts:" + strconv.FormatInt(ownerID, 10)
	return cacher.Of[[]domain.Storefront]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.Storefront, error) {
		return s.repo.List(ctx, ownerID)
	}, cacher.WithExpiry(ctx, cacheTTL))
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.StorefrontInput) (int64, error) {
	if err := s.verifyThemeOwnership(ctx, ownerID, input.ThemeID); err != nil {
		return 0, err
	}
	id, err := s.repo.Create(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}
	s.invalidateCache(ctx, ownerID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, ownerID, id int64, input domain.StorefrontInput) error {
	if err := s.verifyThemeOwnership(ctx, ownerID, input.ThemeID); err != nil {
		return err
	}
	err := s.repo.Update(ctx, ownerID, id, input)
	if err != nil {
		return err
	}
	s.invalidateCache(ctx, ownerID)
	return nil
}

// verifyThemeOwnership ensures a referenced theme belongs to the same owner.
// A zero themeID means no theme is linked, which is allowed.
func (s *Service) verifyThemeOwnership(ctx context.Context, ownerID, themeID int64) error {
	if themeID == 0 {
		return nil
	}
	ok, err := s.repo.ThemeBelongsToOwner(ctx, ownerID, themeID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrThemeNotFound
	}
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, ownerID int64) {
	prefix := "storefronts:" + strconv.FormatInt(ownerID, 10)
	cacher.Of[[]domain.Storefront]().Invalidate(ctx, []string{prefix})
}
