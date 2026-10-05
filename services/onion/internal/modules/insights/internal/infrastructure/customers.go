package infrastructure

import (
	"context"
	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type CustomerKPIs = domain.CustomerKPIs

func (r *Repository) GetCustomerKPIs(ctx context.Context, businessID int64, from, to time.Time) (CustomerKPIs, error) {
	from = dateOnly(from)
	to = dateOnly(to)
	out := CustomerKPIs{}

	total, err := r.queries.CountCustomersAsOf(ctx, sqlc.CountCustomersAsOfParams{
		BusinessID: businessID,
		DATEADD:    to,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: customers total failed", logger.Error(err))
		return CustomerKPIs{}, err
	}
	out.Total = total

	newCustomers, err := r.queries.CountCustomersNewInPeriod(ctx, sqlc.CountCustomersNewInPeriodParams{
		BusinessID: businessID,
		CreatedAt:  dateParam(from),
		DATEADD:    to,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: customers new failed", logger.Error(err))
		return CustomerKPIs{}, err
	}
	out.NewThisPeriod = newCustomers

	active, err := r.queries.CountCustomersActive(ctx, sqlc.CountCustomersActiveParams{
		BusinessID: businessID,
		CreatedAt:  dateParam(from),
		DATEADD:    to,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: customers active failed", logger.Error(err))
		return CustomerKPIs{}, err
	}
	out.Active = active

	churned, err := r.queries.CountCustomersChurned(ctx, sqlc.CountCustomersChurnedParams{
		BusinessID:  businessID,
		CreatedAt:   dateParam(from),
		CreatedAt_2: dateParam(from),
		DATEADD:     to,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: customers churned failed", logger.Error(err))
		return CustomerKPIs{}, err
	}
	out.Churned = churned

	return out, nil
}

func interfaceTime(value interface{}) time.Time {
	switch v := value.(type) {
	case time.Time:
		return v
	case *time.Time:
		if v != nil {
			return *v
		}
	case string:
		if parsed, err := time.Parse(time.RFC3339Nano, v); err == nil {
			return parsed
		}
	case []byte:
		if parsed, err := time.Parse(time.RFC3339Nano, string(v)); err == nil {
			return parsed
		}
	}
	return time.Time{}
}
