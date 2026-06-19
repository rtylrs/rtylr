package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/kpis/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		pool: db,
	}
}

func (r *Repository) GetKPIs(ctx context.Context, ownerID int64) (domain.KPISummary, error) {
	var kpi domain.KPISummary

	// Staff count
	// cross-schema read from hr_staff — active staff headcount for the KPI summary
	staffSQL, staffArgs, err := sq.Select("COUNT(*)").
		From("hr_staff").
		Where(sq.Eq{"owner_id": ownerID, "status": "active"}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build staff count query", logger.Error(err))
	} else {
		err = r.pool.QueryRowContext(ctx, staffSQL, staffArgs...).Scan(&kpi.StaffCount)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Error("Failed to get staff count", logger.Error(err))
		}
	}

	// Customer count
	// cross-schema read from crm_customers — total customers for the KPI summary
	customerSQL, customerArgs, err := sq.Select("COUNT(*)").
		From("crm_customers").
		Where(sq.Eq{"owner_id": ownerID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build customer count query", logger.Error(err))
	} else {
		err = r.pool.QueryRowContext(ctx, customerSQL, customerArgs...).Scan(&kpi.CustomerCount)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Error("Failed to get customer count", logger.Error(err))
		}
	}

	// Low stock count: active ERP products at or below their configured threshold.
	lowStockSQL, lowStockArgs, err := sq.Select("COUNT(*)").
		From("erp_products").
		Where(sq.Eq{
			"owner_id":        ownerID,
			"is_active":       true,
			"track_inventory": true,
		}).
		Where(sq.Expr("low_stock_threshold IS NOT NULL")).
		Where(sq.Expr("stock_quantity <= low_stock_threshold")).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build low stock count query", logger.Error(err))
	} else {
		err = r.pool.QueryRowContext(ctx, lowStockSQL, lowStockArgs...).Scan(&kpi.LowStockCount)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Error("Failed to get low stock count", logger.Error(err))
		}
	}

	// Order count (today)
	// cross-schema read from pos_orders — orders created today for the KPI summary
	orderSQL, orderArgs, err := sq.Select("COUNT(*)").
		From("pos_orders").
		Where(sq.Eq{"owner_id": ownerID}).
		Where(sq.Expr("DATE(created_at) = CURDATE()")).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build order count query", logger.Error(err))
	} else {
		err = r.pool.QueryRowContext(ctx, orderSQL, orderArgs...).Scan(&kpi.OrderCount)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Error("Failed to get order count", logger.Error(err))
		}
	}

	// Revenue (today)
	// cross-schema read from pos_orders — total revenue from orders created today
	revenueSQL, revenueArgs, err := sq.Select("COALESCE(SUM(total), 0)").
		From("pos_orders").
		Where(sq.Eq{"owner_id": ownerID}).
		Where(sq.Expr("DATE(created_at) = CURDATE()")).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build revenue query", logger.Error(err))
	} else {
		err = r.pool.QueryRowContext(ctx, revenueSQL, revenueArgs...).Scan(&kpi.Revenue)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Error("Failed to get revenue", logger.Error(err))
		}
	}

	return kpi, nil
}
