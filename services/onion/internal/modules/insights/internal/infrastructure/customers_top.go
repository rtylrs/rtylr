package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type TopCustomerRow struct {
	CustomerID         int64
	FirstName          string
	LastName           string
	Email              string
	LifetimeSpendMinor int64
	OrderCount         int64
}

func (r *Repository) GetTopCustomers(ctx context.Context, businessID int64, limit int32) ([]TopCustomerRow, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	// cross-schema read from crm_customers + pos_orders — lifetime spend ranking for the top-customers widget
	rows, err := r.queries.GetTopCustomers(ctx, sqlc.GetTopCustomersParams{BusinessID: businessID, Limit: limit})
	if err != nil {
		logger.Ctx(ctx).Error("insights: top customers query failed", logger.Error(err))
		return nil, err
	}
	out := make([]TopCustomerRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, TopCustomerRow{
			CustomerID: row.ID, FirstName: row.FirstName, LastName: row.LastName, Email: row.Email,
			LifetimeSpendMinor: decimalToMinorUnits(row.Spend), OrderCount: row.Orders,
		})
	}
	return out, nil
}
