package infrastructure

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"

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

func (r *Repository) GetTodayCount(ctx context.Context, businessID int64, start, end time.Time) (int, error) {
	var count int
	query, args, err := sq.Select("COUNT(*)").
		From("booking_appointments").
		Where(sq.Eq{"business_id": businessID}).
		Where(sq.GtOrEq{"start_time": start}).
		Where(sq.Lt{"start_time": end}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build today count query", logger.Error(err))
		return 0, err
	}
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get today count in repo", logger.Error(err))
		return 0, err
	}
	return count, nil
}

func (r *Repository) GetUpcomingCount(ctx context.Context, businessID int64, start time.Time) (int, error) {
	var count int
	query, args, err := sq.Select("COUNT(*)").
		From("booking_appointments").
		Where(sq.Eq{"business_id": businessID}).
		Where(sq.GtOrEq{"start_time": start}).
		Where(sq.Eq{"status": []string{"pending", "confirmed"}}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build upcoming count query", logger.Error(err))
		return 0, err
	}
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get upcoming count in repo", logger.Error(err))
		return 0, err
	}
	return count, nil
}

func (r *Repository) GetMonthCounts(ctx context.Context, businessID int64, start time.Time) (total, cancelled int, err error) {
	totalSQL, totalArgs, err := sq.Select("COUNT(*)").
		From("booking_appointments").
		Where(sq.Eq{"business_id": businessID}).
		Where(sq.GtOrEq{"start_time": start}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build total month count query", logger.Error(err))
		return 0, 0, err
	}
	err = r.pool.QueryRowContext(ctx, totalSQL, totalArgs...).Scan(&total)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get total month count in repo", logger.Error(err))
		return 0, 0, err
	}

	cancelledSQL, cancelledArgs, err := sq.Select("COUNT(*)").
		From("booking_appointments").
		Where(sq.Eq{"business_id": businessID}).
		Where(sq.GtOrEq{"start_time": start}).
		Where(sq.Eq{"status": "cancelled"}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build cancelled month count query", logger.Error(err))
		return 0, 0, err
	}
	err = r.pool.QueryRowContext(ctx, cancelledSQL, cancelledArgs...).Scan(&cancelled)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get cancelled month count in repo", logger.Error(err))
		return 0, 0, err
	}

	return total, cancelled, nil
}

func (r *Repository) GetMonthRevenue(ctx context.Context, businessID int64, start time.Time) (float64, error) {
	var revenue sql.NullFloat64
	query, args, err := sq.Select("COALESCE(SUM(bs.price), 0)").
		From("booking_appointments ba").
		Join("booking_services bs ON ba.service_id = bs.id").
		Where(sq.Eq{"ba.business_id": businessID}).
		Where(sq.GtOrEq{"ba.start_time": start}).
		Where(sq.Eq{"ba.status": []string{"completed", "confirmed"}}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build revenue query", logger.Error(err))
		return 0, err
	}
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&revenue)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get revenue in repo", logger.Error(err))
		return 0, err
	}
	if revenue.Valid {
		return revenue.Float64, nil
	}
	return 0, nil
}
