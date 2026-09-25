package infrastructure

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/platform/logger"
)

type ChurnRateRow struct {
	WindowDays   int32
	ChurnedCount int64
	ActiveCount  int64
	ChurnPct     float64
}

// GetChurnRates classifies customers against the requested period: customers
// whose most recent paid order before the period ended are churned, while those
// ordering inside the period are active.
//
// dynamic-query: squirrel
func (r *Repository) GetChurnRates(ctx context.Context, businessID int64, from, to time.Time) ([]ChurnRateRow, error) {
	from = dateOnly(from)
	to = dateOnly(to)
	periodEnd := to.AddDate(0, 0, 1)
	inner := sq.
		Select("c.id", "MAX(o.created_at) AS last_order_at").
		From("crm_customers c").
		LeftJoin("pos_orders o ON o.customer_id = c.id AND o.business_id = c.business_id AND o.payment_status = 'paid' AND o.created_at < ?", periodEnd).
		Where(sq.Eq{"c.business_id": businessID}).
		GroupBy("c.id").
		Having("MAX(o.created_at) IS NOT NULL")

	query := sq.Select().
		FromSelect(inner, "t").
		Column(sq.Expr("COALESCE(SUM(CASE WHEN last_order_at >= ? THEN 1 ELSE 0 END), 0)", from)).
		Column(sq.Expr("COALESCE(SUM(CASE WHEN last_order_at < ? THEN 1 ELSE 0 END), 0)", from))

	q, args, err := query.PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("insights: build churn query failed", logger.Error(err))
		return nil, err
	}

	var active, churned int64
	if err := r.pool.QueryRowContext(ctx, q, args...).Scan(&active, &churned); err != nil {
		logger.Ctx(ctx).Error("insights: churn query failed", logger.Error(err))
		return nil, err
	}

	total := active + churned
	var pct float64
	if total > 0 {
		pct = float64(churned) / float64(total) * 100
	}
	windowDays := int32(to.Sub(from).Hours()/24 + 1)
	return []ChurnRateRow{{
		WindowDays:   windowDays,
		ChurnedCount: churned,
		ActiveCount:  active,
		ChurnPct:     pct,
	}}, nil
}
