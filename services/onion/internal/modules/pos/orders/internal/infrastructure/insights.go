package infrastructure

import (
	"context"
	"database/sql"
	"time"

	poscontract "github.com/voxire/onion/internal/contracts/pos"
	"github.com/voxire/onion/internal/modules/pos/orders/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

// The reads below serve Insights reports, so Insights never queries POS
// tables itself. Customer figures use the customer recorded on each paid
// order. Period arguments named from/to on customer reads are calendar days.

// InsightsNetRevenue sums paid and refunded orders paid in the range, net of
// their refunds.
func (r *Repository) InsightsNetRevenue(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	total, err := r.queries.InsightsNetRevenue(ctx, sqlc.InsightsNetRevenueParams{
		BusinessID: businessID,
		DateFrom:   sql.NullTime{Time: from, Valid: true},
		DateTo:     sql.NullTime{Time: to, Valid: true},
	})
	if err != nil {
		return 0, err
	}
	return sqlutil.DecimalToMinorUnits(total), nil
}

// InsightsOrdersToday counts orders created today and sums their totals in
// major units, both by the database's calendar day.
func (r *Repository) InsightsOrdersToday(ctx context.Context, businessID int64) (int64, float64, error) {
	count, err := r.queries.InsightsCountOrdersToday(ctx, businessID)
	if err != nil {
		return 0, 0, err
	}
	revenue, err := r.queries.InsightsSumRevenueToday(ctx, businessID)
	if err != nil {
		return 0, 0, err
	}
	return count, sqlutil.DecimalToFloat(revenue), nil
}

// InsightsBuyerCount counts customers with a paid order in the period.
func (r *Repository) InsightsBuyerCount(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	return r.queries.InsightsCountBuyers(ctx, sqlc.InsightsCountBuyersParams{
		BusinessID: businessID,
		DateFrom:   sql.NullTime{Time: from, Valid: true},
		DateTo:     to,
	})
}

// InsightsLapsedBuyerCount counts customers who paid for an order before the
// period but not during it.
func (r *Repository) InsightsLapsedBuyerCount(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	return r.queries.InsightsCountLapsedBuyers(ctx, sqlc.InsightsCountLapsedBuyersParams{
		BusinessID: businessID,
		DateFrom:   sql.NullTime{Time: from, Valid: true},
		DateTo:     to,
	})
}

func (r *Repository) InsightsTopBuyers(ctx context.Context, businessID int64, query poscontract.TopBuyersQuery) ([]poscontract.CustomerPurchases, error) {
	byOrderCount := int64(0)
	if query.ByOrderCount {
		byOrderCount = 1
	}
	direction := int64(-1)
	if query.Descending {
		direction = 1
	}
	rows, err := r.queries.InsightsTopBuyers(ctx, sqlc.InsightsTopBuyersParams{
		BusinessID:   businessID,
		DateFrom:     sql.NullTime{Time: query.From, Valid: true},
		DateTo:       query.To,
		ByOrderCount: byOrderCount,
		Direction:    direction,
		Limit:        query.Limit,
		Offset:       query.Offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]poscontract.CustomerPurchases, 0, len(rows))
	for _, row := range rows {
		name, _ := row.CustomerName.(string)
		if raw, ok := row.CustomerName.([]byte); ok {
			name = string(raw)
		}
		out = append(out, poscontract.CustomerPurchases{
			CustomerID:      row.CustomerID.Int64,
			CustomerName:    name,
			SpendMinorUnits: sqlutil.DecimalToMinorUnits(row.Spend),
			Orders:          row.Orders,
			LastOrderAt:     insightsTime(row.LastOrderAt),
		})
	}
	return out, nil
}

// InsightsBuyerSpendBuckets buckets customers with a paid order by lifetime
// paid spend.
func (r *Repository) InsightsBuyerSpendBuckets(ctx context.Context, businessID int64) ([]poscontract.SpendBucket, error) {
	rows, err := r.queries.InsightsBuyerSpendBuckets(ctx, businessID)
	if err != nil {
		return nil, err
	}
	out := make([]poscontract.SpendBucket, 0, len(rows))
	for _, row := range rows {
		out = append(out, poscontract.SpendBucket{
			Label:           row.Label,
			LowerMinorUnits: int64(row.LowerMinorUnits),
			UpperMinorUnits: int64(row.UpperMinorUnits),
			CustomerCount:   row.CustomerCount,
		})
	}
	return out, nil
}

// InsightsPaidProductSales totals paid, non-cancelled line items per product
// for orders created from `from`, and up to `to` unless it is zero.
func (r *Repository) InsightsPaidProductSales(ctx context.Context, businessID int64, from, to time.Time) ([]poscontract.ProductSales, error) {
	rows, err := r.queries.InsightsPaidProductSales(ctx, sqlc.InsightsPaidProductSalesParams{
		BusinessID: businessID,
		DateFrom:   sql.NullTime{Time: from, Valid: true},
		DateTo:     sql.NullTime{Time: to, Valid: !to.IsZero()},
	})
	if err != nil {
		return nil, err
	}
	out := make([]poscontract.ProductSales, 0, len(rows))
	for _, row := range rows {
		out = append(out, poscontract.ProductSales{
			ProductID:         row.ProductID.Int64,
			Units:             sqlutil.DecimalToFloat(row.Units),
			RevenueMinorUnits: sqlutil.DecimalToMinorUnits(row.Revenue),
		})
	}
	return out, nil
}

// InsightsLifetimeProductSales totals every line item ever sold per product,
// whatever the order or item status. Products with no net quantity are left
// out.
func (r *Repository) InsightsLifetimeProductSales(ctx context.Context, businessID int64) ([]poscontract.ProductSales, error) {
	rows, err := r.queries.InsightsLifetimeProductSales(ctx, businessID)
	if err != nil {
		return nil, err
	}
	out := make([]poscontract.ProductSales, 0, len(rows))
	for _, row := range rows {
		out = append(out, poscontract.ProductSales{
			ProductID:         row.ProductID.Int64,
			Units:             sqlutil.DecimalToFloat(row.Units),
			RevenueMinorUnits: row.RevenueMinorUnits,
		})
	}
	return out, nil
}

// InsightsProductsSoldSince lists products on any order created since `since`.
func (r *Repository) InsightsProductsSoldSince(ctx context.Context, businessID int64, since time.Time) ([]int64, error) {
	rows, err := r.queries.InsightsProductsSoldSince(ctx, sqlc.InsightsProductsSoldSinceParams{
		BusinessID: businessID,
		CreatedAt:  sql.NullTime{Time: since, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	out := make([]int64, 0, len(rows))
	for _, id := range rows {
		out = append(out, id.Int64)
	}
	return out, nil
}

// InsightsLastChange returns the latest order update or refund.
func (r *Repository) InsightsLastChange(ctx context.Context, businessID int64) (time.Time, error) {
	return r.queries.InsightsPOSLastChange(ctx, sqlc.InsightsPOSLastChangeParams{BusinessID: businessID})
}

func insightsTime(value any) time.Time {
	switch v := value.(type) {
	case time.Time:
		return v
	case []byte:
		return parseInsightsTime(string(v))
	case string:
		return parseInsightsTime(v)
	}
	return time.Time{}
}

func parseInsightsTime(value string) time.Time {
	for _, layout := range []string{time.RFC3339Nano, time.DateTime} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed
		}
	}
	return time.Time{}
}
