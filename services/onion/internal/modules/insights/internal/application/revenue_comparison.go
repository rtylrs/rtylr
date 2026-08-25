package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// RevenueComparisonPayload is the outer envelope returned by the
// comparison endpoint. Delta is reported as a float percentage so the
// frontend can render "+12.4%" without doing the division itself.
type RevenueComparisonPayload struct {
	ThisPeriodMinorUnits int64   `json:"this_period"`
	LastPeriodMinorUnits int64   `json:"last_period"`
	DeltaPct             float64 `json:"delta_pct"`
}

// GetRevenueComparison resolves the window, fetches this+prior totals, and
// computes a delta percentage safely even when last period is zero.
func (s *Service) GetRevenueComparison(ctx context.Context, businessID int64, from, to time.Time) (RevenueComparisonPayload, error) {
	key, err := s.insightsCacheKey(ctx, businessID, fmt.Sprintf("revenue:cmp:%s:%s", from.Format("2006-01-02"), to.Format("2006-01-02")))
	if err != nil {
		return RevenueComparisonPayload{}, err
	}
	out, err := cacher.Of[RevenueComparisonPayload]().GetOrComputeSingle(ctx, key, func(ctx context.Context) (RevenueComparisonPayload, error) {
		row, rerr := s.repo.GetRevenueComparison(ctx, businessID, from, to)
		if rerr != nil {
			return RevenueComparisonPayload{}, rerr
		}
		var delta float64
		if row.LastPeriodMinor > 0 {
			delta = (float64(row.ThisPeriodMinor-row.LastPeriodMinor) / float64(row.LastPeriodMinor)) * 100
		}
		return RevenueComparisonPayload{
			ThisPeriodMinorUnits: row.ThisPeriodMinor,
			LastPeriodMinorUnits: row.LastPeriodMinor,
			DeltaPct:             delta,
		}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return RevenueComparisonPayload{}, err
	}
	return out, nil
}
