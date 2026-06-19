package infrastructure

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/platform/logger"
)

// CustomerKPIs is the denormalised payload the service layer hands back.
// Keeping it flat here avoids the repository needing to know about the
// frontend's JSON contract.
type CustomerKPIs struct {
	Total        int64
	NewThisMonth int64
	Active       int64
	Churned      int64
}

// GetCustomerKPIs computes four cohort-style counters over the CRM customer
// table, using pos_orders.created_at as the "last activity" signal.
//
// "Active" = at least one POS order in the last 30 days.
// "Churned" = has orders historically but none in the last 90+ days. This
// is a loose heuristic; a strict churn model would need a signup date
// filter, which sits in the dedicated /customers/churn endpoint.
func (r *Repository) GetCustomerKPIs(ctx context.Context, businessID int64) (CustomerKPIs, error) {
	now := nowFn()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	cutoffActive := now.AddDate(0, 0, -30)
	cutoffChurn := now.AddDate(0, 0, -90)

	out := CustomerKPIs{}

	// cross-schema read from crm_customers — total customer count for the customers KPI card
	totalSQL, totalArgs, err := sq.
		Select("COUNT(*)").
		From("crm_customers").
		Where(sq.Eq{"business_id": businessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("insights: build customers total query failed", logger.Error(err))
		return CustomerKPIs{}, err
	}
	if err := r.pool.QueryRowContext(ctx, totalSQL, totalArgs...).Scan(&out.Total); err != nil {
		logger.Ctx(ctx).Error("insights: customers total failed", logger.Error(err))
		return CustomerKPIs{}, err
	}

	// cross-schema read from crm_customers — new-this-month for the customers KPI card
	newSQL, newArgs, err := sq.
		Select("COUNT(*)").
		From("crm_customers").
		Where(sq.Eq{"business_id": businessID}).
		Where(sq.GtOrEq{"created_at": monthStart}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("insights: build customers new query failed", logger.Error(err))
		return CustomerKPIs{}, err
	}
	if err := r.pool.QueryRowContext(ctx, newSQL, newArgs...).Scan(&out.NewThisMonth); err != nil {
		logger.Ctx(ctx).Error("insights: customers new failed", logger.Error(err))
		return CustomerKPIs{}, err
	}

	// cross-schema read from crm_customers + pos_orders — active (ordered in last 30d)
	activeSQL, activeArgs, err := sq.
		Select("COUNT(DISTINCT c.id)").
		From("crm_customers c").
		Join("pos_orders o ON o.customer_id = c.id AND o.business_id = c.business_id").
		Where(sq.Eq{"c.business_id": businessID}).
		Where(sq.GtOrEq{"o.created_at": cutoffActive}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("insights: build customers active query failed", logger.Error(err))
		return CustomerKPIs{}, err
	}
	if err := r.pool.QueryRowContext(ctx, activeSQL, activeArgs...).Scan(&out.Active); err != nil {
		logger.Ctx(ctx).Error("insights: customers active failed", logger.Error(err))
		return CustomerKPIs{}, err
	}

	// cross-schema read from crm_customers + pos_orders — churned (had orders, none in 90d+)
	churnSQL, churnArgs, err := sq.
		Select("COUNT(*)").
		From("crm_customers c").
		Where(sq.Eq{"c.business_id": businessID}).
		Where(sq.Expr("EXISTS (SELECT 1 FROM pos_orders o WHERE o.business_id = c.business_id AND o.customer_id = c.id)")).
		Where(sq.Expr("NOT EXISTS (SELECT 1 FROM pos_orders o2 WHERE o2.business_id = c.business_id AND o2.customer_id = c.id AND o2.created_at >= ?)", cutoffChurn)).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("insights: build customers churned query failed", logger.Error(err))
		return CustomerKPIs{}, err
	}
	if err := r.pool.QueryRowContext(ctx, churnSQL, churnArgs...).Scan(&out.Churned); err != nil {
		logger.Ctx(ctx).Error("insights: customers churned failed", logger.Error(err))
		return CustomerKPIs{}, err
	}

	return out, nil
}
