package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/modules/insights/kpis/internal/domain"
	"github.com/voxire/onion/internal/modules/insights/kpis/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
	}
}

func (r *Repository) GetKPIs(ctx context.Context, businessID int64) (domain.KPISummary, error) {
	var kpi domain.KPISummary
	// A failed sub-query must surface as an error: otherwise the zero-filled
	// summary is indistinguishable from "no activity" and gets cached as such.
	var queryErrors []error

	// cross-schema read from hr_staff — active staff headcount for the KPI summary
	if count, err := r.queries.CountActiveStaff(ctx, businessID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Error("Failed to get staff count", logger.Error(err))
			queryErrors = append(queryErrors, fmt.Errorf("staff count: %w", err))
		}
	} else {
		kpi.StaffCount = count
	}

	// cross-schema read from crm_customers — total customers for the KPI summary
	if count, err := r.queries.CountCustomers(ctx, businessID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Error("Failed to get customer count", logger.Error(err))
			queryErrors = append(queryErrors, fmt.Errorf("customer count: %w", err))
		}
	} else {
		kpi.CustomerCount = count
	}

	if count, err := r.queries.CountLowStockProducts(ctx, businessID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Error("Failed to get low stock count", logger.Error(err))
			queryErrors = append(queryErrors, fmt.Errorf("low stock count: %w", err))
		}
	} else {
		kpi.LowStockCount = count
	}

	// cross-schema read from pos_orders — orders created today for the KPI summary
	if count, err := r.queries.CountOrdersToday(ctx, businessID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Error("Failed to get order count", logger.Error(err))
			queryErrors = append(queryErrors, fmt.Errorf("order count: %w", err))
		}
	} else {
		kpi.OrderCount = count
	}

	// cross-schema read from pos_orders — total revenue from orders created today
	if revenue, err := r.queries.SumRevenueToday(ctx, businessID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Error("Failed to get revenue", logger.Error(err))
			queryErrors = append(queryErrors, fmt.Errorf("revenue: %w", err))
		}
	} else {
		kpi.Revenue = sqlutil.DecimalToFloat(revenue)
	}

	return kpi, errors.Join(queryErrors...)
}
