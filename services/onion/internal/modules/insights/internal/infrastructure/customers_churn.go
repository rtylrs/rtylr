package infrastructure

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/platform/logger"
)

// ChurnRateRow captures a single rolling-window churn rate: of the
// customers who had ever placed an order, what fraction have been silent
// for the entire window [now-windowDays, now].
type ChurnRateRow struct {
	WindowDays   int32
	ChurnedCount int64
	ActiveCount  int64
	ChurnPct     float64
}

// Standard windows used by the frontend churn panel.
var churnWindows = []int32{30, 60, 90, 180}

// GetChurnRates returns churn percentages for each of the standard
// windows. A fresh tenant with no orders at all returns zero-valued
// rows rather than an error, per the fallback rules in CLAUDE.md §8.
//
// dynamic-query: squirrel
func (r *Repository) GetChurnRates(ctx context.Context, businessID int64) ([]ChurnRateRow, error) {
	now := nowFn()
	// cross-schema read from crm_customers + pos_orders — rolling churn windows for the customer health panel
	//
	// Compute every window from one last-order aggregation. The previous
	// implementation reran the same customer/order scan once per window.
	inner := sq.
		Select("c.id", "MAX(o.created_at) AS last_order_at").
		From("crm_customers c").
		LeftJoin("pos_orders o ON o.customer_id = c.id AND o.business_id = c.business_id").
		Where(sq.Eq{"c.business_id": businessID}).
		GroupBy("c.id").
		Having("MAX(o.created_at) IS NOT NULL")
	query := sq.Select().FromSelect(inner, "t")
	for _, days := range churnWindows {
		cutoff := now.AddDate(0, 0, -int(days))
		query = query.
			Column(sq.Expr("SUM(CASE WHEN last_order_at < ? THEN 1 ELSE 0 END)", cutoff)).
			Column(sq.Expr("SUM(CASE WHEN last_order_at >= ? THEN 1 ELSE 0 END)", cutoff))
	}
	q, args, err := query.PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("insights: build churn query failed", logger.Error(err))
		return nil, err
	}

	counts := make([]int64, len(churnWindows)*2)
	scanArgs := make([]any, len(counts))
	for i := range counts {
		scanArgs[i] = &nullableInt64{&counts[i]}
	}
	if err := r.pool.QueryRowContext(ctx, q, args...).Scan(scanArgs...); err != nil {
		logger.Ctx(ctx).Error("insights: churn query failed", logger.Error(err))
		return nil, err
	}

	out := make([]ChurnRateRow, 0, len(churnWindows))
	for i, days := range churnWindows {
		churned, active := counts[i*2], counts[i*2+1]
		total := churned + active
		var pct float64
		if total > 0 {
			pct = (float64(churned) / float64(total)) * 100
		}
		out = append(out, ChurnRateRow{
			WindowDays:   days,
			ChurnedCount: churned,
			ActiveCount:  active,
			ChurnPct:     pct,
		})
	}
	return out, nil
}

// nullableInt64 is a sql.Scanner that folds NULL -> 0. Needed because the
// SUM(...)/GROUP BY subquery emits NULL when the inner set is empty.
type nullableInt64 struct{ Target *int64 }

func (n *nullableInt64) Scan(src interface{}) error {
	if src == nil {
		*n.Target = 0
		return nil
	}
	switch v := src.(type) {
	case int64:
		*n.Target = v
	case []byte:
		// decimal bytes; truncate fractional part — counts are whole.
		var acc int64
		for _, b := range v {
			if b < '0' || b > '9' {
				break
			}
			acc = acc*10 + int64(b-'0')
		}
		*n.Target = acc
	case float64:
		*n.Target = int64(v)
	case time.Time:
		*n.Target = 0
	default:
		*n.Target = 0
	}
	return nil
}
