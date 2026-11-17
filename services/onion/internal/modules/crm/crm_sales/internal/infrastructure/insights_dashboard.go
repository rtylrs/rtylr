package infrastructure

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/money"
)

// OpenDealPipeline counts the business's open deals and sums their
// first-year value.
func (r *Repository) OpenDealPipeline(ctx context.Context, businessID int64) (int64, money.Amount, error) {
	row, err := r.queries.OpenDealPipeline(ctx, businessID)
	if err != nil {
		return 0, 0, err
	}
	return row.Deals, amount(row.Amount), nil
}

// WonDealsBetween counts deals won with closed_at in [from, to] and sums
// their first-year value.
func (r *Repository) WonDealsBetween(ctx context.Context, businessID int64, from, to time.Time) (int64, money.Amount, error) {
	row, err := r.queries.WonDealsBetween(ctx, sqlc.WonDealsBetweenParams{
		BusinessID: businessID, DateFrom: someTime(from), DateTo: someTime(to),
	})
	if err != nil {
		return 0, 0, err
	}
	return row.Deals, amount(row.Amount), nil
}
