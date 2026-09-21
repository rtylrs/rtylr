package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/apps/internal/domain"
	"github.com/voxire/onion/internal/modules/apps/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const appsCacheTTL = 5 * time.Minute

type repository interface {
	GetEnabledAddons(ctx context.Context, businessID int64) (map[string]bool, error)
	GetUserAllowedApps(ctx context.Context, businessID, userID int64) ([]string, error)
	GetUserBusinessID(ctx context.Context, userID int64) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func (s *Service) ListApps() []domain.App {
	return domain.AvailableApps()
}

// GetEnabledAddons keeps its fail-closed signature: callers get an empty map
// when the lookup fails, so addon-gated apps stay disabled. The fetch closure
// propagates the error though, so a transient failure is never cached as "no
// addons enabled" for the whole TTL.
func (s *Service) GetEnabledAddons(ctx context.Context, businessID int64) map[string]bool {
	enabled, err := cacher.Of[map[string]bool]().GetOrComputeSingle(ctx,
		cacheKeys.BusinessEnabledAddons(businessID),
		func(ctx context.Context) (map[string]bool, error) {
			return s.repo.GetEnabledAddons(ctx, businessID)
		},
		cacher.WithExpiry(ctx, appsCacheTTL),
	)
	if err != nil {
		return map[string]bool{}
	}
	if enabled == nil {
		return map[string]bool{}
	}
	return enabled
}

// GetUserAllowedApps keeps its fail-closed signature for the same reason: no
// allowed apps means no app access. The fetch closure propagates the error so a
// transient failure is not cached as an empty app list.
func (s *Service) GetUserAllowedApps(ctx context.Context, businessID, userID int64) []string {
	apps, err := cacher.Of[[]string]().GetOrComputeSingle(ctx,
		cacheKeys.UserAllowedApps(userID, businessID),
		func(ctx context.Context) ([]string, error) {
			return s.repo.GetUserAllowedApps(ctx, businessID, userID)
		},
		cacher.WithExpiry(ctx, appsCacheTTL),
	)
	if err != nil || apps == nil {
		return []string{}
	}
	return apps
}

func (s *Service) GetVisibleAppsForUser(ctx context.Context, businessID, userID int64, role string) ([]domain.App, error) {
	if userID == 0 {
		return s.GetVisibleApps(ctx, businessID, userID, role), nil
	}

	resolvedBusinessID, err := s.resolveBusinessID(ctx, businessID, userID)
	if err != nil {
		return nil, err
	}
	return s.GetVisibleApps(ctx, resolvedBusinessID, userID, role), nil
}

func (s *Service) resolveBusinessID(ctx context.Context, businessID, userID int64) (int64, error) {
	if businessID > 0 {
		return businessID, nil
	}
	return cacher.Of[int64]().GetOrComputeSingle(ctx,
		cacheKeys.UserBusinessID(userID),
		func(ctx context.Context) (int64, error) {
			return s.repo.GetUserBusinessID(ctx, userID)
		},
		cacher.WithExpiry(ctx, appsCacheTTL),
	)
}

// GetVisibleApps returns the app catalogue projected for a specific user:
// addon-gated apps are toggled on only when the business has the addon enabled
// and (for non-owners/admins) the user has explicit access.
//
// Pass an empty role and userID==0 for anonymous callers — only ungated core
// apps are returned in that case.
func (s *Service) GetVisibleApps(ctx context.Context, businessID, userID int64, role string) []domain.App {
	all := s.ListApps()

	if userID == 0 {
		out := make([]domain.App, 0, len(all))
		for _, a := range all {
			if a.AddonGated == "" {
				out = append(out, a)
			}
		}
		return out
	}

	enabledAddons := s.GetEnabledAddons(ctx, businessID)
	project := func(a domain.App) domain.App {
		if a.AddonGated != "" {
			a.IsEnabled = enabledAddons[a.AddonGated]
		}
		return a
	}

	if hasFullCatalogAccess(role) {
		out := make([]domain.App, 0, len(all))
		for _, a := range all {
			out = append(out, project(a))
		}
		return out
	}

	allowedApps := s.GetUserAllowedApps(ctx, businessID, userID)
	allowed := make(map[string]bool, len(allowedApps))
	for _, id := range allowedApps {
		allowed[id] = true
	}

	out := make([]domain.App, 0, len(all))
	for _, a := range all {
		if allowed[a.ID] {
			out = append(out, project(a))
		}
	}
	return out
}

func hasFullCatalogAccess(role string) bool {
	switch role {
	case "owner", "admin", "console_admin":
		return true
	default:
		return false
	}
}
