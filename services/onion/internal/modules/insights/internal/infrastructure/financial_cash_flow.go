package infrastructure

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type CashFlowDayRow = domain.CashFlowDayRow

func (r *Repository) GetCashFlowSeries(ctx context.Context, businessID int64, from, to time.Time) ([]CashFlowDayRow, error) {
	inflow, err := r.sources.Finance.InsightsPaidInvoicesByDay(ctx, businessID, from, to)
	if err != nil {
		logger.Ctx(ctx).Error("insights: daily paid invoices failed", logger.Error(err))
		return nil, err
	}
	outflow, err := r.sources.Finance.InsightsExpensesByDay(ctx, businessID, from, to)
	if err != nil {
		logger.Ctx(ctx).Error("insights: daily expenses failed", logger.Error(err))
		return nil, err
	}
	out := make([]CashFlowDayRow, 0)
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		out = append(out, CashFlowDayRow{
			Date: key, InflowMinorUnits: inflow[key], OutflowMinorUnits: outflow[key],
		})
	}
	return out, nil
}
