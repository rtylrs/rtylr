package infrastructure

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type CustomerKPIs = domain.CustomerKPIs

// GetCustomerKPIs reports CRM's customer counts alongside POS purchase
// activity. Active and churned count the customers recorded on paid orders.
func (r *Repository) GetCustomerKPIs(ctx context.Context, businessID int64, from, to time.Time) (CustomerKPIs, error) {
	from = dateOnly(from)
	to = dateOnly(to)
	out := CustomerKPIs{}
	var err error
	if out.Total, err = r.sources.CRM.InsightsCustomersAsOf(ctx, businessID, to); err != nil {
		logger.Ctx(ctx).Error("insights: customers total failed", logger.Error(err))
		return CustomerKPIs{}, err
	}
	if out.NewThisPeriod, err = r.sources.CRM.InsightsNewCustomers(ctx, businessID, from, to); err != nil {
		logger.Ctx(ctx).Error("insights: customers new failed", logger.Error(err))
		return CustomerKPIs{}, err
	}
	if out.Active, err = r.sources.POS.InsightsBuyerCount(ctx, businessID, from, to); err != nil {
		logger.Ctx(ctx).Error("insights: customers active failed", logger.Error(err))
		return CustomerKPIs{}, err
	}
	if out.Churned, err = r.sources.POS.InsightsLapsedBuyerCount(ctx, businessID, from, to); err != nil {
		logger.Ctx(ctx).Error("insights: customers churned failed", logger.Error(err))
		return CustomerKPIs{}, err
	}
	return out, nil
}
