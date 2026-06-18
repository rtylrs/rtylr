package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/booking_dashboard/internal/domain"
	"github.com/voxire/onion/internal/modules/booking_dashboard/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"golang.org/x/sync/errgroup"
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

	var (
		todayCount     int
		upcomingCount  int
		totalMonth     int
		cancelledMonth int
		revenue        float64
	)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		var err error
		todayCount, err = s.repo.GetTodayCount(groupCtx, ownerID, startOfDay, endOfDay)
		return err
	})
	group.Go(func() error {
		var err error
		upcomingCount, err = s.repo.GetUpcomingCount(groupCtx, ownerID, endOfDay)
		return err
	})
	group.Go(func() error {
		var err error
		totalMonth, cancelledMonth, err = s.repo.GetMonthCounts(groupCtx, ownerID, startOfMonth)
		return err
	})
	group.Go(func() error {
		var err error
		revenue, err = s.repo.GetMonthRevenue(groupCtx, ownerID, startOfMonth)
		return err
	})
	if err := group.Wait(); err != nil {
		return domain.DashboardData{}, err
	}

	data := domain.DashboardData{
		TodayCount:    todayCount,
		UpcomingCount: upcomingCount,
		Revenue:       revenue,
	}
	if totalMonth > 0 {
		data.CancellationRate = float64(cancelledMonth) / float64(totalMonth) * 100
	}

	return data, nil
}
