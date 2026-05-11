package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/kpis/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		pool: dbclient.Pool(),
	}
}

func (r *Repository) GetKPIs(ctx context.Context, ownerID int64) (domain.KPISummary, error) {
	var kpi domain.KPISummary

	// Staff count
	err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM user_app_permissions WHERE owner_id = ? AND is_active = TRUE`, ownerID,
	).Scan(&kpi.StaffCount)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("Failed to get staff count", logger.Error(err))
	}

	// Customer count
	err = r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM customers WHERE owner_id = ?`, ownerID,
	).Scan(&kpi.CustomerCount)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("Failed to get customer count", logger.Error(err))
	}

	// Low stock count
	err = r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM products WHERE owner_id = ? AND quantity <= reorder_level AND quantity > 0`, ownerID,
	).Scan(&kpi.LowStockCount)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("Failed to get low stock count", logger.Error(err))
	}

	// Order count (today)
	err = r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM pos_orders WHERE owner_id = ? AND DATE(created_at) = CURDATE()`, ownerID,
	).Scan(&kpi.OrderCount)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("Failed to get order count", logger.Error(err))
	}

	// Revenue (today)
	err = r.pool.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(total), 0) FROM pos_orders WHERE owner_id = ? AND DATE(created_at) = CURDATE()`, ownerID,
	).Scan(&kpi.Revenue)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("Failed to get revenue", logger.Error(err))
	}

	return kpi, nil
}
