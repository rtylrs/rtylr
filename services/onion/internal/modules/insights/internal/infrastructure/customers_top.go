package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type TopCustomerRow = domain.TopCustomerRow

type TopCustomerQuery = domain.TopCustomerQuery

type TopCustomerPage = domain.TopCustomerPage

func (r *Repository) GetTopCustomers(ctx context.Context, businessID int64, query TopCustomerQuery) (TopCustomerPage, error) {
	query.From = dateOnly(query.From)
	query.To = dateOnly(query.To)
	if query.Limit <= 0 {
		query.Limit = 10
	}
	if query.Limit > 100 {
		query.Limit = 100
	}
	if query.Offset < 0 {
		query.Offset = 0
	}

	total, err := r.queries.CountCustomersWithPaidOrders(ctx, sqlc.CountCustomersWithPaidOrdersParams{
		BusinessID: businessID,
		CreatedAt:  dateParam(query.From),
		DATEADD:    query.To,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: count top customers failed", logger.Error(err))
		return TopCustomerPage{}, err
	}

	byOrderCount := int64(0)
	if query.ByOrderCount {
		byOrderCount = 1
	}
	direction := int64(-1)
	if query.Descending {
		direction = 1
	}
	rows, err := r.queries.GetTopCustomers(ctx, sqlc.GetTopCustomersParams{
		BusinessID:   businessID,
		CreatedAt:    dateParam(query.From),
		DATEADD:      query.To,
		ByOrderCount: byOrderCount,
		Direction:    direction,
		Limit:        query.Limit,
		Offset:       query.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: top customers query failed", logger.Error(err))
		return TopCustomerPage{}, err
	}

	out := make([]TopCustomerRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, TopCustomerRow{
			CustomerID:         row.ID,
			FirstName:          row.FirstName,
			LastName:           row.LastName,
			Email:              row.Email,
			LifetimeSpendMinor: decimalToMinorUnits(row.Spend),
			OrderCount:         row.Orders,
			LastOrderAt:        interfaceTime(row.LastOrderAt),
		})
	}
	return TopCustomerPage{Rows: out, Total: total}, nil
}
