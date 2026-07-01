package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type CustomerKPIs struct {
	Total        int64
	NewThisMonth int64
	Active       int64
	Churned      int64
}

func (r *Repository) GetCustomerKPIs(ctx context.Context, businessID int64) (CustomerKPIs, error) {
	now := nowFn()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	cutoffActive := now.AddDate(0, 0, -30)
	cutoffChurn := now.AddDate(0, 0, -90)

	out := CustomerKPIs{}

	// cross-schema read from crm_customers — total customer count for the customers KPI card
	total, err := r.queries.CountCustomersTotal(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: customers total failed", logger.Error(err))
		return CustomerKPIs{}, err
	}
	out.Total = total

	// cross-schema read from crm_customers — new-this-month for the customers KPI card
	n, err := r.queries.CountCustomersNewThisMonth(ctx, sqlc.CountCustomersNewThisMonthParams{
		BusinessID: businessID,
		CreatedAt:  sql.NullTime{Time: monthStart, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: customers new failed", logger.Error(err))
		return CustomerKPIs{}, err
	}
	out.NewThisMonth = n

	// cross-schema read from crm_customers + pos_orders — active (ordered in last 30d)
	n, err = r.queries.CountCustomersActive(ctx, sqlc.CountCustomersActiveParams{
		BusinessID: businessID,
		CreatedAt:  sql.NullTime{Time: cutoffActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: customers active failed", logger.Error(err))
		return CustomerKPIs{}, err
	}
	out.Active = n

	// cross-schema read from crm_customers + pos_orders — churned (had orders, none in 90d+)
	n, err = r.queries.CountCustomersChurned(ctx, sqlc.CountCustomersChurnedParams{
		BusinessID: businessID,
		CreatedAt:  sql.NullTime{Time: cutoffChurn, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: customers churned failed", logger.Error(err))
		return CustomerKPIs{}, err
	}
	out.Churned = n

	return out, nil
}
