package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/apps/internal/domain"
	"github.com/voxire/onion/internal/modules/apps/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const appsCacheTTL = 5 * time.Minute

type repository interface {
	GetEnabledAddons(ctx context.Context, businessID int64) (map[string]bool, error)
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

func (s *Service) GetVisibleAppsForUser(ctx context.Context, businessID, userID int64, access *authorization.EffectiveAccess) ([]domain.App, error) {
	if userID == 0 {
		return s.GetVisibleApps(ctx, businessID, userID, nil), nil
	}
	if businessID <= 0 && access != nil {
		businessID = access.BusinessID
	}

	resolvedBusinessID, err := s.resolveBusinessID(ctx, businessID, userID)
	if err != nil {
		return nil, err
	}
	allowedApps := make([]string, 0)
	if access != nil {
		for _, app := range access.Apps {
			if app.Licensed && app.HasAccess {
				allowedApps = append(allowedApps, string(app.AppID))
			}
		}
	}
	return s.GetVisibleApps(ctx, resolvedBusinessID, userID, allowedApps), nil
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

// GetVisibleApps projects the catalog using the server-resolved app grants.
// Addon-gated apps are enabled only when the business has the addon enabled.
func (s *Service) GetVisibleApps(ctx context.Context, businessID, userID int64, allowedApps []string) []domain.App {
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
