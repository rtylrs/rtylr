package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/pos_day_summary/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_day_summary/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) GetDaySummary(ctx context.Context, businessID int64, date string) (domain.DaySummary, error) {
	var summary domain.DaySummary

	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}
	t, parseErr := time.Parse("2006-01-02", date)
	if parseErr != nil {
		logger.Ctx(ctx).Error("Failed to parse date for GetDaySummary", logger.Error(parseErr))
		return summary, parseErr
	}
	end := t.AddDate(0, 0, 1)
	row, err := r.queries.GetDaySummaryByDate(ctx, sqlc.GetDaySummaryByDateParams{
		BusinessID: businessID,
		PaidAt:     sql.NullTime{Time: t, Valid: true},
		PaidAtEnd:  sql.NullTime{Time: end, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get day summary", logger.Error(err))
		return summary, err
	}
	summary.TotalSales = interfaceToFloat(row.TotalSales)
	summary.TotalTax = interfaceToFloat(row.TotalTax)
	summary.TotalTips = interfaceToFloat(row.TotalTips)
	summary.TotalOrders = int(row.TotalOrders)

	refundTotal, err := r.queries.GetRefundTotalByDate(ctx, sqlc.GetRefundTotalByDateParams{
		BusinessID:    businessID,
		RefundedAt:    sql.NullTime{Time: t, Valid: true},
		RefundedAtEnd: sql.NullTime{Time: end, Valid: true},
	})
	if err != nil {
		// A financial report must not silently report zero refunds.
		logger.Ctx(ctx).Error("Failed to get refund total", logger.Error(err))
		return summary, err
	}
	summary.TotalRefunds = interfaceToFloat(refundTotal)

	breakdown, err := r.queries.GetPaymentBreakdownByDate(ctx, sqlc.GetPaymentBreakdownByDateParams{
		BusinessID: businessID,
		PaidAt:     sql.NullTime{Time: t, Valid: true},
		PaidAtEnd:  sql.NullTime{Time: end, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get payment breakdown", logger.Error(err))
		return summary, err
	}
	summary.PaymentBreakdown = make(map[string]float64, len(breakdown))
	for _, b := range breakdown {
		summary.PaymentBreakdown[string(b.PaymentMethod.PosOrdersPaymentMethod)] = interfaceToFloat(b.MethodTotal)
	}

	return summary, nil
}

func interfaceToFloat(v any) float64 {
	if v == nil {
		return 0
	}
	if f, ok := v.(float64); ok {
		return f
	}
	var str string
	switch val := v.(type) {
	case []byte:
		str = string(val)
	case string:
		str = val
	default:
		str = fmt.Sprintf("%v", val)
	}
	if str == "" {
		return 0
	}
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		logger.Global().Warn("failed to parse day-summary decimal from database",
			logger.String("value", str), logger.Error(err))
		return 0
	}
	return f
}
