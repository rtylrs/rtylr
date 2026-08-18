package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/booking_dashboard/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
	}
}

func (r *Repository) GetTodayCount(ctx context.Context, businessID int64, start, end time.Time) (int, error) {
	cnt, err := r.queries.CountAppointmentsToday(ctx, sqlc.CountAppointmentsTodayParams{
		BusinessID:  businessID,
		StartTime:   start,
		StartTime_2: end,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get today count in repo", logger.Error(err))
		return 0, err
	}
	return int(cnt), nil
}

func (r *Repository) GetUpcomingCount(ctx context.Context, businessID int64, start time.Time) (int, error) {
	cnt, err := r.queries.CountUpcomingAppointments(ctx, sqlc.CountUpcomingAppointmentsParams{
		BusinessID: businessID,
		StartTime:  start,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get upcoming count in repo", logger.Error(err))
		return 0, err
	}
	return int(cnt), nil
}

func (r *Repository) GetMonthCounts(ctx context.Context, businessID int64, start time.Time) (total, cancelled int, err error) {
	total64, err := r.queries.CountAppointmentsThisMonth(ctx, sqlc.CountAppointmentsThisMonthParams{
		BusinessID: businessID,
		StartTime:  start,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get total month count in repo", logger.Error(err))
		return 0, 0, err
	}

	cancelled64, err := r.queries.CountCancelledAppointmentsThisMonth(ctx, sqlc.CountCancelledAppointmentsThisMonthParams{
		BusinessID: businessID,
		StartTime:  start,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get cancelled month count in repo", logger.Error(err))
		return 0, 0, err
	}

	return int(total64), int(cancelled64), nil
}

func (r *Repository) GetMonthRevenue(ctx context.Context, businessID int64, start time.Time) (money.Amount, error) {
	revenue, err := r.queries.SumMonthRevenue(ctx, sqlc.SumMonthRevenueParams{
		BusinessID: businessID,
		StartTime:  start,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get revenue in repo", logger.Error(err))
		return 0, err
	}
	return interfaceToMoney(revenue), nil
}

func interfaceToMoney(v interface{}) money.Amount {
	switch val := v.(type) {
	case float64:
		parsed, err := money.FromFloat64(val)
		if err != nil {
			return 0
		}
		return parsed
	case int64:
		return parseMoneyString(strconv.FormatInt(val, 10))
	case []byte:
		return parseMoneyString(string(val))
	case string:
		return parseMoneyString(val)
	default:
		return parseMoneyString(fmt.Sprintf("%v", val))
	}
}

func parseMoneyString(s string) money.Amount {
	value, err := money.FromString(s)
	if err != nil {
		return 0
	}
	return value
}
