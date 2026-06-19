package domain

import (
	"context"
	"time"
)

type Repository interface {
	GetTodayCount(ctx context.Context, businessID int64, start, end time.Time) (int, error)
	GetUpcomingCount(ctx context.Context, businessID int64, start time.Time) (int, error)
	GetMonthCounts(ctx context.Context, businessID int64, start time.Time) (total, cancelled int, err error)
	GetMonthRevenue(ctx context.Context, businessID int64, start time.Time) (float64, error)
}

type DashboardData struct {
	TodayCount       int     `json:"today_count"`
	UpcomingCount    int     `json:"upcoming_count"`
	CancellationRate float64 `json:"cancellation_rate"`
	Revenue          float64 `json:"revenue"`
}
