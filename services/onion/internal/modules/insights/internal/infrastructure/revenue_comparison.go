package infrastructure

import (
	"context"
	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"time"
)

// RevenueComparisonRow captures the "this period vs last period" pair for
// the revenue comparison endpoint. The repository returns raw minor-unit
// totals; the service layer computes the delta percentage.
type RevenueComparisonRow = domain.RevenueComparisonRow

// GetRevenueComparison sums paid invoice revenue for [from, to] and also
// for the preceding window of the same length, so the frontend can render
// a "this vs last" bar pair without asking the backend for two windows.
func (r *Repository) GetRevenueComparison(ctx context.Context, businessID int64, from, to time.Time) (RevenueComparisonRow, error) {
	windowDays := int(to.Sub(from).Hours() / 24)
	priorTo := from.AddDate(0, 0, -1)
	priorFrom := priorTo.AddDate(0, 0, -windowDays)

	thisAmt, err := sumPaidInvoicesMinorUnits(ctx, r.queries, businessID, from, to)
	if err != nil {
		return RevenueComparisonRow{}, err
	}
	lastAmt, err := sumPaidInvoicesMinorUnits(ctx, r.queries, businessID, priorFrom, priorTo)
	if err != nil {
		return RevenueComparisonRow{}, err
	}
	return RevenueComparisonRow{
		ThisPeriodMinor: thisAmt,
		LastPeriodMinor: lastAmt,
	}, nil
}
