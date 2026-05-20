package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/booking_dashboard/internal/domain"
	"github.com/voxire/onion/internal/modules/booking_dashboard/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const bookingDashboardTTL = 1 * time.Minute

type Service struct {
	repo domain.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func (s *Service) GetDashboard(ctx context.Context, ownerID int64) (domain.DashboardData, error) {
	return cacher.Of[domain.DashboardData]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped("booking_dashboard", ownerID, "data"),
		func(ctx context.Context) (domain.DashboardData, error) {
			return s.fetchDashboard(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, bookingDashboardTTL),
	)
}

func (s *Service) fetchDashboard(ctx context.Context, ownerID int64) (domain.DashboardData, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var data domain.DashboardData

	todayCount, err := s.repo.GetTodayCount(ctx, ownerID, startOfDay, endOfDay)
	if err != nil {
		return data, err
	}
	data.TodayCount = todayCount

	upcomingCount, err := s.repo.GetUpcomingCount(ctx, ownerID, endOfDay)
	if err != nil {
		return data, err
	}
	data.UpcomingCount = upcomingCount

	totalMonth, cancelledMonth, err := s.repo.GetMonthCounts(ctx, ownerID, startOfMonth)
	if err != nil {
		return data, err
	}

	if totalMonth > 0 {
		data.CancellationRate = float64(cancelledMonth) / float64(totalMonth) * 100
	}

	revenue, err := s.repo.GetMonthRevenue(ctx, ownerID, startOfMonth)
	if err != nil {
		return data, err
	}
	data.Revenue = revenue

	return data, nil
}
