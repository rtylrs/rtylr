package infrastructure

import (
	"context"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/platform/logger"
)

// TopCustomerRow is one entry in the "top N by lifetime spend" ranking.
// Spend is summed from pos_orders.total (paid-status rows only) so the
// total lines up with what the customer has actually paid, not what they
// may have an open invoice for.
type TopCustomerRow struct {
	CustomerID         int64
	FirstName          string
	LastName           string
	Email              string
	LifetimeSpendMinor int64
	OrderCount         int64
}

// GetTopCustomers returns up to `limit` customers by lifetime POS spend.
// A limit of 0 is coerced upstream; we cap at 100 to prevent abuse, in
// line with the pagination cap documented in onion/CLAUDE.md §4.
func (r *Repository) GetTopCustomers(ctx context.Context, businessID int64, limit int32) ([]TopCustomerRow, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	// cross-schema read from crm_customers + pos_orders — lifetime spend ranking for the top-customers widget
	q, args, err := sq.
		Select(
			"c.id",
			"COALESCE(c.first_name, '')",
			"COALESCE(c.last_name, '')",
			"COALESCE(c.email, '')",
			"COALESCE(SUM(o.total), 0) AS spend",
			"COUNT(o.id) AS orders",
		).
		From("crm_customers c").
		LeftJoin("pos_orders o ON o.customer_id = c.id AND o.business_id = c.business_id AND o.payment_status = 'paid'").
		Where(sq.Eq{"c.business_id": businessID}).
		GroupBy("c.id", "c.first_name", "c.last_name", "c.email").
		OrderBy("spend DESC").
		Limit(uint64(limit)).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("insights: build top customers query failed", logger.Error(err))
		return nil, err
	}
	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("insights: top customers query failed", logger.Error(err))
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Warn("insights: rows.Close failed", logger.Error(closeErr))
		}
	}()

	out := make([]TopCustomerRow, 0, limit)
	for rows.Next() {
		var row TopCustomerRow
		var spend interface{}
		if scanErr := rows.Scan(&row.CustomerID, &row.FirstName, &row.LastName, &row.Email, &spend, &row.OrderCount); scanErr != nil {
			logger.Ctx(ctx).Error("insights: top customers scan failed", logger.Error(scanErr))
			return nil, scanErr
		}
		row.LifetimeSpendMinor = decimalToMinorUnits(spend)
		out = append(out, row)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	return out, nil
}
