package domain

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/money"
)

type Repository interface {
	GetTodayCount(ctx context.Context, businessID int64, start, end time.Time) (int, error)
	GetUpcomingCount(ctx context.Context, businessID int64, start time.Time) (int, error)
	GetMonthCounts(ctx context.Context, businessID int64, start time.Time) (total, cancelled int, err error)
	GetMonthRevenue(ctx context.Context, businessID int64, start time.Time) (money.Amount, error)
}

type DashboardData struct {
	TodayCount       int
	UpcomingCount    int
	CancellationRate float64
	Revenue          money.Amount
}
