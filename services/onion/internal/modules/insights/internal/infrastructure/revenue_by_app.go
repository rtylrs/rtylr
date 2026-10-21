package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type RevenueByAppRow = domain.RevenueByAppRow

func (r *Repository) GetRevenueByApp(ctx context.Context, businessID int64) ([]RevenueByAppRow, error) {
	thisFrom, thisTo, _, _ := mtdRanges(nowFn())

	posTotal, err := r.sources.POS.InsightsNetRevenue(ctx, businessID, thisFrom, thisTo)
	if err != nil {
		logger.Ctx(ctx).Error("insights: revenue by app (pos) failed", logger.Error(err))
		return nil, err
	}

	finTotal, err := r.sources.Finance.InsightsInvoiceRevenue(ctx, businessID, thisFrom, thisTo)
	if err != nil {
		logger.Ctx(ctx).Error("insights: revenue by app (finance) failed", logger.Error(err))
		return nil, err
	}

	return []RevenueByAppRow{
		{App: "pos", AmountMinorUnits: posTotal},
		{App: "finance", AmountMinorUnits: finTotal},
		{App: "erp", AmountMinorUnits: 0},
		{App: "crm", AmountMinorUnits: 0},
	}, nil
}
