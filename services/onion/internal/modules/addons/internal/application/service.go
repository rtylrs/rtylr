package application

import (
	"context"
	"database/sql"
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

var hiddenCatalogAddons = map[string]struct{}{
	"ai_chat": {},
}

type repository interface {
	ListAll(ctx context.Context) ([]domain.Addon, error)
	ListForBusiness(ctx context.Context, businessID int64) ([]domain.BusinessAddon, error)
	HasActiveGrant(ctx context.Context, businessID int64, addonID string) (bool, error)
	Enable(ctx context.Context, businessID int64, addonID string) error
	Disable(ctx context.Context, businessID int64, addonID string) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func invalidateAddonsScope(ctx context.Context, businessID int64) {
	// Apps visibility depends on enabled addons.
	cacher.InvalidateScope(ctx,
		cacheKeys.ScopedPrefix(addonsScope, businessID),
		cacheKeys.BusinessPrefix(businessID),
	)
}

func isVisibleCatalogAddon(addonID string) bool {
	_, hidden := hiddenCatalogAddons[addonID]
	return !hidden
}

func visibleCatalogAddons(addons []domain.Addon) []domain.Addon {
	result := make([]domain.Addon, 0, len(addons))
	for _, addon := range addons {
		if isVisibleCatalogAddon(addon.ID) {
			result = append(result, addon)
		}
	}
	return result
}

func visibleBusinessAddons(addons []domain.BusinessAddon) []domain.BusinessAddon {
	result := make([]domain.BusinessAddon, 0, len(addons))
	for _, addon := range addons {
		if isVisibleCatalogAddon(addon.AddonID) {
			result = append(result, addon)
		}
	}
	return result
}

// ListAll returns the addon catalog. Global, so cached under a fixed key
// with a long TTL; invalidated only by a process restart or manual flush.
func (s *Service) ListAll(ctx context.Context) ([]domain.Addon, error) {
	addons, err := cacher.Of[[]domain.Addon]().GetOrComputeSingle(ctx, addonsCatalog,
		func(ctx context.Context) ([]domain.Addon, error) {
			return s.repo.ListAll(ctx)
		},
		cacher.WithExpiry(ctx, 1*time.Hour),
	)
	if err != nil {
		return nil, err
	}
	return visibleCatalogAddons(addons), nil
}

func (s *Service) ListForBusiness(ctx context.Context, businessID int64) ([]domain.BusinessAddon, error) {
	addons, err := cacher.Of[[]domain.BusinessAddon]().GetOrComputeSingle(ctx,
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
			enabledMap := make(map[string]domain.BusinessAddon, len(enabled))
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
	if err != nil {
		return nil, err
	}
	return visibleBusinessAddons(addons), nil
}

func (s *Service) Enable(ctx context.Context, businessID int64, addonID string) error {
	addons, err := s.repo.ListAll(ctx)
	if err != nil {
		return err
	}
	var requested *domain.Addon
	for i := range addons {
		if addons[i].ID == addonID {
			requested = &addons[i]
			break
		}
	}
	if requested == nil || !isVisibleCatalogAddon(addonID) {
		return domain.ErrAddonNotFound
	}
	if !requested.IsFree || requested.Price > 0 {
		entitled, grantErr := s.repo.HasActiveGrant(ctx, businessID, addonID)
		if grantErr != nil {
			return grantErr
		}
		if !entitled {
			return domain.ErrAddonNotEntitled
		}
	}
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
