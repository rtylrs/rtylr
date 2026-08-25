package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type RevenueByAppRow struct {
	App              string
	AmountMinorUnits int64
}

func (r *Repository) GetRevenueByApp(ctx context.Context, businessID int64) ([]RevenueByAppRow, error) {
	thisFrom, thisTo, _, _ := mtdRanges(nowFn())

	// cross-schema read from pos_orders — POS-tracked paid order revenue for the split donut
	posTotal, err := r.queries.SumPosRevenueMTD(ctx, sqlc.SumPosRevenueMTDParams{
		BusinessID: businessID,
		PaidAt:     sql.NullTime{Time: thisFrom, Valid: true},
		PaidAt_2:   sql.NullTime{Time: thisTo, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: revenue by app (pos) failed", logger.Error(err))
		return nil, err
	}

	// cross-schema read from finance_invoices — one-off invoice revenue for the split donut
	finTotal, err := r.queries.SumFinanceRevenueMTD(ctx, sqlc.SumFinanceRevenueMTDParams{
		BusinessID:      businessID,
		ReportingDate:   sql.NullTime{Time: thisFrom, Valid: true},
		ReportingDate_2: sql.NullTime{Time: thisTo, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: revenue by app (finance) failed", logger.Error(err))
		return nil, err
	}

	return []RevenueByAppRow{
		{App: "pos", AmountMinorUnits: decimalToMinorUnits(posTotal)},
		{App: "finance", AmountMinorUnits: decimalToMinorUnits(finTotal)},
		{App: "erp", AmountMinorUnits: 0},
		{App: "crm", AmountMinorUnits: 0},
	}, nil
}
