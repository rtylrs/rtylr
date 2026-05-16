package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/addons/internal/domain"
	"github.com/voxire/onion/internal/modules/addons/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	addonsScope   = "addons"
	addonsTTL     = 5 * time.Minute
	addonsCatalog = "cache:addons:catalog"
)

type repository interface {
	ListAll(ctx context.Context) ([]domain.Addon, error)
	ListForBusiness(ctx context.Context, businessID int64) ([]domain.BusinessAddon, error)
	Enable(ctx context.Context, businessID int64, addonID string) error
	Disable(ctx context.Context, businessID int64, addonID string) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func invalidateAddonsScope(ctx context.Context, businessID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(addonsScope, businessID))
		// Apps visibility depends on enabled addons.
		cache.DeleteByPattern(ctx, cacheKeys.BusinessPrefix(businessID))
	}
}

// ListAll returns the addon catalog. Global, so cached under a fixed key
// with a long TTL; invalidated only by a process restart or manual flush.
func (s *Service) ListAll(ctx context.Context) ([]domain.Addon, error) {
	return cacher.Of[[]domain.Addon]().GetOrComputeSingle(ctx, addonsCatalog,
		func(ctx context.Context) ([]domain.Addon, error) {
			return s.repo.ListAll(ctx)
		},
		cacher.WithExpiry(ctx, 1*time.Hour),
	)
}

func (s *Service) ListForBusiness(ctx context.Context, businessID int64) ([]domain.BusinessAddon, error) {
	return cacher.Of[[]domain.BusinessAddon]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(addonsScope, businessID, "merged"),
		func(ctx context.Context) ([]domain.BusinessAddon, error) {
			all, err := s.repo.ListAll(ctx)
			if err != nil {
				return nil, err
			}
			enabled, err := s.repo.ListForBusiness(ctx, businessID)
			if err != nil {
				return nil, err
			}
			enabledMap := make(map[string]domain.BusinessAddon)
			for _, ba := range enabled {
				enabledMap[ba.AddonID] = ba
			}
			result := make([]domain.BusinessAddon, 0, len(all))
			for _, addon := range all {
				if ba, ok := enabledMap[addon.ID]; ok {
					result = append(result, ba)
				} else {
					result = append(result, domain.BusinessAddon{
						AddonID:     addon.ID,
						Name:        addon.Name,
						Description: addon.Description,
						Icon:        addon.Icon,
						Price:       addon.Price,
						IsFree:      addon.IsFree,
						IsEnabled:   false,
					})
				}
			}
			return result, nil
		},
		cacher.WithExpiry(ctx, addonsTTL),
	)
}

func (s *Service) Enable(ctx context.Context, businessID int64, addonID string) error {
	if err := s.repo.Enable(ctx, businessID, addonID); err != nil {
		return err
	}
	invalidateAddonsScope(ctx, businessID)
	return nil
}

func (s *Service) Disable(ctx context.Context, businessID int64, addonID string) error {
	if err := s.repo.Disable(ctx, businessID, addonID); err != nil {
		return err
	}
	invalidateAddonsScope(ctx, businessID)
	return nil
}

func (s *Service) IsEnabled(ctx context.Context, businessID int64, addonID string) bool {
	addons, err := s.ListForBusiness(ctx, businessID)
	if err != nil {
		return false
	}
	for _, a := range addons {
		if a.AddonID == addonID && a.IsEnabled {
			return true
		}
	}
	return false
}
