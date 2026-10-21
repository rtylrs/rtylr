package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/modules/insights/kpis/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// Sources are the owning modules the KPI summary is read from. The summary
// never queries their tables.
type Sources struct {
	Staff interface {
		InsightsActiveStaff(context.Context, int64) (int64, error)
	}
	Customers interface {
		InsightsCustomerCount(context.Context, int64) (int64, error)
	}
	Products interface {
		InsightsLowStockCount(context.Context, int64) (int64, error)
	}
	Orders interface {
		InsightsOrdersToday(context.Context, int64) (int64, float64, error)
	}
}

type Repository struct {
	sources Sources
}

func NewRepository(sources Sources) *Repository {
	return &Repository{sources: sources}
}

func (r *Repository) GetKPIs(ctx context.Context, businessID int64) (domain.KPISummary, error) {
	var kpi domain.KPISummary
	// A failed sub-query must surface as an error: otherwise the zero-filled
	// summary is indistinguishable from "no activity" and gets cached as such.
	var queryErrors []error
	fail := func(label string, err error) bool {
		if err == nil || errors.Is(err, sql.ErrNoRows) {
			return err != nil
		}
		logger.Ctx(ctx).Error("Failed to get "+label, logger.Error(err))
		queryErrors = append(queryErrors, fmt.Errorf("%s: %w", label, err))
		return true
	}

	if count, err := r.sources.Staff.InsightsActiveStaff(ctx, businessID); !fail("staff count", err) {
		kpi.StaffCount = count
	}
	if count, err := r.sources.Customers.InsightsCustomerCount(ctx, businessID); !fail("customer count", err) {
		kpi.CustomerCount = count
	}
	if count, err := r.sources.Products.InsightsLowStockCount(ctx, businessID); !fail("low stock count", err) {
		kpi.LowStockCount = count
	}
	if count, revenue, err := r.sources.Orders.InsightsOrdersToday(ctx, businessID); !fail("order count", err) {
		kpi.OrderCount = count
		kpi.Revenue = revenue
	}

	return kpi, errors.Join(queryErrors...)
}
