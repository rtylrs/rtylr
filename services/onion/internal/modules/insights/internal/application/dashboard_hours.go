package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// GetHoursSummary returns the Work time card for the overview, cached like the
// other dashboard cards.
func (s *Service) GetHoursSummary(ctx context.Context, businessID int64) (domain.HoursSummary, error) {
	key, err := s.insightsCacheKey(ctx, businessID, "dashboard-hours")
	if err != nil {
		return domain.HoursSummary{}, err
	}
	return cacher.Of[domain.HoursSummary]().GetOrComputeSingle(ctx, key+"|domain=v1", func(ctx context.Context) (domain.HoursSummary, error) {
		return s.readHoursSummary(ctx, businessID, time.Now())
	}, cacher.WithExpiry(ctx, dashboardCacheTTL))
}

func (s *Service) readHoursSummary(ctx context.Context, businessID int64, now time.Time) (domain.HoursSummary, error) {
	flow := s.dashboard.Flow
	if flow == nil {
		return domain.HoursSummary{}, errDashboardReadersUnavailable
	}
	thisFrom, thisTo, priorFrom, priorTo := dashboardMTDRanges(now)

	current, err := flow.FlowLoggedTime(ctx, businessID, thisFrom, thisTo)
	if err != nil {
		return domain.HoursSummary{}, fmt.Errorf("read current logged time: %w", err)
	}
	prior, err := flow.FlowLoggedTime(ctx, businessID, priorFrom, priorTo)
	if err != nil {
		return domain.HoursSummary{}, fmt.Errorf("read previous logged time: %w", err)
	}
	return domain.HoursSummary{
		Logged:       domain.CountCard{Value: current.TotalMinutes, DeltaPct: dashboardDeltaPct(current.TotalMinutes, prior.TotalMinutes)},
		PriorMinutes: prior.TotalMinutes,
		Projects:     publicProjectHours(current.Projects),
	}, nil
}

// publicProjectHours names only the projects the overview's Project progress
// list may show: public and not archived. Insights viewers are not
// necessarily members of a private project, so its name stays out.
func publicProjectHours(in []domain.FlowProjectTime) []domain.ProjectHours {
	out := make([]domain.ProjectHours, 0, len(in))
	for _, p := range in {
		if p.Visibility != "public" || p.Status == "archived" || p.Minutes <= 0 {
			continue
		}
		out = append(out, domain.ProjectHours{ProjectID: p.ProjectID, Name: p.Name, Minutes: p.Minutes})
	}
	return out
}
