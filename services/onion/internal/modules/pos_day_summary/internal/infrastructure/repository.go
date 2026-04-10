package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/pos_day_summary/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) GetDaySummary(ctx context.Context, ownerID int64, date string) (domain.DaySummary, error) {
	var summary domain.DaySummary

	dateFilter := "CURDATE()"
	args := []interface{}{ownerID}
	if date != "" {
		dateFilter = "?"
		args = append(args, date)
	}

	query := `SELECT
		COALESCE(SUM(total), 0),
		COALESCE(SUM(tax_amount), 0),
		COUNT(*)
	FROM pos_orders
	WHERE owner_id = ? AND payment_status = 'paid' AND DATE(paid_at) = ` + dateFilter

	err := r.pool.QueryRowContext(ctx, query, args...).Scan(
		&summary.TotalSales, &summary.TotalTax, &summary.TotalOrders,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get day summary", logger.Error(err))
		return summary, err
	}

	// Refunds
	refundArgs := []interface{}{ownerID}
	refundFilter := "CURDATE()"
	if date != "" {
		refundFilter = "?"
		refundArgs = append(refundArgs, date)
	}

	refundQuery := `SELECT COALESCE(SUM(amount), 0) FROM pos_refunds
		WHERE owner_id = ? AND DATE(refunded_at) = ` + refundFilter

	err = r.pool.QueryRowContext(ctx, refundQuery, refundArgs...).Scan(&summary.TotalRefunds)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get refund total", logger.Error(err))
	}

	// Payment breakdown
	breakdownArgs := []interface{}{ownerID}
	breakdownFilter := "CURDATE()"
	if date != "" {
		breakdownFilter = "?"
		breakdownArgs = append(breakdownArgs, date)
	}

	breakdownQuery := `SELECT payment_method, COALESCE(SUM(total), 0)
		FROM pos_orders
		WHERE owner_id = ? AND payment_status = 'paid' AND DATE(paid_at) = ` + breakdownFilter + `
		GROUP BY payment_method`

	rows, err := r.pool.QueryContext(ctx, breakdownQuery, breakdownArgs...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get payment breakdown", logger.Error(err))
		return summary, err
	}
	defer rows.Close()

	summary.PaymentBreakdown = make(map[string]float64)
	for rows.Next() {
		var method string
		var amount float64
		if err := rows.Scan(&method, &amount); err != nil {
			continue
		}
		summary.PaymentBreakdown[method] = amount
	}

	return summary, nil
}
