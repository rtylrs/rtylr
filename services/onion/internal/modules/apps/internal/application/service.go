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
	GetEnabledAddons(ctx context.Context, businessID int64) map[string]bool
	GetUserAllowedApps(ctx context.Context, businessID, userID int64) []string
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

func (s *Service) GetEnabledAddons(ctx context.Context, businessID int64) map[string]bool {
	return s.repo.GetEnabledAddons(ctx, businessID)
}

func (s *Service) GetUserAllowedApps(ctx context.Context, businessID, userID int64) []string {
	return s.repo.GetUserAllowedApps(ctx, businessID, userID)
}

// GetVisibleApps returns the app catalogue projected for a specific user:
// addon-gated apps are toggled on only when the business has the addon enabled
// and (for non-owners/admins) the user has explicit access. Result is cached
// per-user; team/permission/addon changes invalidate `cache:user:{id}:*`.
//
// Pass an empty role and userID==0 for anonymous callers — only ungated core
// apps are returned in that case.
func (s *Service) GetVisibleApps(ctx context.Context, businessID, userID int64, role string) []domain.App {
	if userID == 0 {
		return s.computeVisibleApps(ctx, 0, 0, "")
	}
	apps, _ := cacher.Of[[]domain.App]().GetOrComputeSingle(
		ctx,
		cacheKeys.UserApps(userID),
		func(ctx context.Context) ([]domain.App, error) {
			return s.computeVisibleApps(ctx, businessID, userID, role), nil
		},
		cacher.WithExpiry(ctx, appsCacheTTL),
	)
	return apps
}

func (s *Service) computeVisibleApps(ctx context.Context, businessID, userID int64, role string) []domain.App {
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

	if role == "owner" || role == "admin" {
		out := make([]domain.App, 0, len(all))
		for _, a := range all {
			out = append(out, project(a))
		}
		return out
	}

	allowed := make(map[string]bool)
	for _, id := range s.GetUserAllowedApps(ctx, businessID, userID) {
		allowed[id] = true
	}

	out := make([]domain.App, 0)
	for _, a := range all {
		if allowed[a.ID] {
			out = append(out, project(a))
		}
	}
	return out
}
