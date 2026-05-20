package infrastructure

import (
	"context"
	"database/sql"
	"time"

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

func (r *Repository) GetTodayCount(ctx context.Context, ownerID int64, start, end time.Time) (int, error) {
	var count int
	err := r.pool.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM booking_appointments WHERE owner_id = ? AND start_time >= ? AND start_time < ?",
		ownerID, start, end).Scan(&count)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get today count in repo", logger.Error(err))
		return 0, err
	}
	return count, nil
}

func (r *Repository) GetUpcomingCount(ctx context.Context, ownerID int64, start time.Time) (int, error) {
	var count int
	err := r.pool.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM booking_appointments WHERE owner_id = ? AND start_time >= ? AND status IN ('pending','confirmed')",
		ownerID, start).Scan(&count)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get upcoming count in repo", logger.Error(err))
		return 0, err
	}
	return count, nil
}

func (r *Repository) GetMonthCounts(ctx context.Context, ownerID int64, start time.Time) (total, cancelled int, err error) {
	err = r.pool.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM booking_appointments WHERE owner_id = ? AND start_time >= ?",
		ownerID, start).Scan(&total)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get total month count in repo", logger.Error(err))
		return 0, 0, err
	}

	err = r.pool.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM booking_appointments WHERE owner_id = ? AND start_time >= ? AND status = 'cancelled'",
		ownerID, start).Scan(&cancelled)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get cancelled month count in repo", logger.Error(err))
		return 0, 0, err
	}

	return total, cancelled, nil
}

func (r *Repository) GetMonthRevenue(ctx context.Context, ownerID int64, start time.Time) (float64, error) {
	var revenue sql.NullFloat64
	err := r.pool.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(bs.price), 0) FROM booking_appointments ba
		 JOIN booking_services bs ON ba.service_id = bs.id
		 WHERE ba.owner_id = ? AND ba.start_time >= ? AND ba.status IN ('completed','confirmed')`,
		ownerID, start).Scan(&revenue)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get revenue in repo", logger.Error(err))
		return 0, err
	}
	if revenue.Valid {
		return revenue.Float64, nil
	}
	return 0, nil
}
