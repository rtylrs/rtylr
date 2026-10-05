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
	out, err := cacher.Of[RevenueComparisonPayload]().GetOrComputeSingle(ctx, key+"|domain=v2", func(ctx context.Context) (RevenueComparisonPayload, error) {
		if s.dashboard.Finance == nil {
			return RevenueComparisonPayload{}, errDashboardReadersUnavailable
		}
		windowDays := int(to.Sub(from).Hours() / 24)
		priorTo := from.AddDate(0, 0, -1)
		priorFrom := priorTo.AddDate(0, 0, -windowDays)
		thisAmount, rerr := s.dashboard.Finance.InsightsPaidInvoiceTotal(ctx, businessID, from, to)
		if rerr != nil {
			return RevenueComparisonPayload{}, fmt.Errorf("read current paid invoice total: %w", rerr)
		}
		lastAmount, rerr := s.dashboard.Finance.InsightsPaidInvoiceTotal(ctx, businessID, priorFrom, priorTo)
		if rerr != nil {
			return RevenueComparisonPayload{}, fmt.Errorf("read previous paid invoice total: %w", rerr)
		}
		var delta float64
		if lastAmount > 0 {
			delta = (float64(thisAmount-lastAmount) / float64(lastAmount)) * 100
		}
		return RevenueComparisonPayload{
			ThisPeriodMinorUnits: thisAmount,
			LastPeriodMinorUnits: lastAmount,
			DeltaPct:             delta,
		}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return RevenueComparisonPayload{}, err
	}
	return out, nil
}
