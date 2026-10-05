package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// dashboardCacheTTL keeps the expensive cross-schema rollup behind a
// 5-minute cache, per the task spec. Mutations in downstream modules
// (invoices, orders, customers) do not busts this key — acceptable for
// an analytics surface where 5-minute freshness is the product requirement.
const dashboardCacheTTL = 5 * time.Minute

// GetDashboard returns the four headline cards for the given owner.
// Result is cached per-owner under `insights:{businessID}:dashboard`.
func (s *Service) GetDashboard(ctx context.Context, businessID int64) (domain.DashboardSnapshot, error) {
	key, err := s.insightsCacheKey(ctx, businessID, "dashboard")
	if err != nil {
		return domain.DashboardSnapshot{}, err
	}
	return cacher.Of[domain.DashboardSnapshot]().GetOrComputeSingle(ctx, key+"|domain=v2", func(ctx context.Context) (domain.DashboardSnapshot, error) {
		return s.repo.GetDashboard(ctx, businessID, domain.DashboardInput{Now: time.Now()})
	}, cacher.WithExpiry(ctx, dashboardCacheTTL))
}
