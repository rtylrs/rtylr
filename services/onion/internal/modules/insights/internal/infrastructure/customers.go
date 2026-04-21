package infrastructure

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

// CustomerKPIs is the denormalised payload the service layer hands back.
// Keeping it flat here avoids the repository needing to know about the
// frontend's JSON contract.
type CustomerKPIs struct {
	Total       int64
	NewThisMonth int64
	Active      int64
	Churned     int64
}

// GetCustomerKPIs computes four cohort-style counters over the CRM customer
// table, using pos_orders.created_at as the "last activity" signal.
//
// "Active" = at least one POS order in the last 30 days.
// "Churned" = has orders historically but none in the last 90+ days. This
// is a loose heuristic; a strict churn model would need a signup date
// filter, which sits in the dedicated /customers/churn endpoint.
func (r *Repository) GetCustomerKPIs(ctx context.Context, ownerID int64) (CustomerKPIs, error) {
	now := nowFn()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	cutoffActive := now.AddDate(0, 0, -30)
	cutoffChurn := now.AddDate(0, 0, -90)

	out := CustomerKPIs{}

	// cross-schema read from crm_customers — total customer count for the customers KPI card
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM crm_customers WHERE owner_id = ?`,
		ownerID,
	).Scan(&out.Total); err != nil {
		logger.Ctx(ctx).Error("insights: customers total failed", logger.Error(err))
		return CustomerKPIs{}, err
	}

	// cross-schema read from crm_customers — new-this-month for the customers KPI card
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM crm_customers WHERE owner_id = ? AND created_at >= ?`,
		ownerID, monthStart,
	).Scan(&out.NewThisMonth); err != nil {
		logger.Ctx(ctx).Error("insights: customers new failed", logger.Error(err))
		return CustomerKPIs{}, err
	}

	// cross-schema read from crm_customers + pos_orders — active (ordered in last 30d)
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT c.id)
		 FROM crm_customers c
		 JOIN pos_orders o ON o.customer_id = c.id AND o.owner_id = c.owner_id
		 WHERE c.owner_id = ? AND o.created_at >= ?`,
		ownerID, cutoffActive,
	).Scan(&out.Active); err != nil {
		logger.Ctx(ctx).Error("insights: customers active failed", logger.Error(err))
		return CustomerKPIs{}, err
	}

	// cross-schema read from crm_customers + pos_orders — churned (had orders, none in 90d+)
	const churnQ = `
		SELECT COUNT(*)
		FROM crm_customers c
		WHERE c.owner_id = ?
		  AND EXISTS (SELECT 1 FROM pos_orders o WHERE o.owner_id = c.owner_id AND o.customer_id = c.id)
		  AND NOT EXISTS (SELECT 1 FROM pos_orders o2 WHERE o2.owner_id = c.owner_id AND o2.customer_id = c.id AND o2.created_at >= ?)
	`
	if err := r.pool.QueryRowContext(ctx, churnQ, ownerID, cutoffChurn).Scan(&out.Churned); err != nil {
		logger.Ctx(ctx).Error("insights: customers churned failed", logger.Error(err))
		return CustomerKPIs{}, err
	}

	return out, nil
}
