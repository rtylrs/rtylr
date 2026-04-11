package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/booking_dashboard/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Service struct{ pool *sql.DB }

func NewService(_ context.Context) *Service {
	return &Service{pool: dbclient.Pool()}
}

func (s *Service) GetDashboard(ctx context.Context, ownerID int64) (domain.DashboardData, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var data domain.DashboardData

	err := s.pool.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM booking_appointments WHERE owner_id = ? AND start_time >= ? AND start_time < ?",
		ownerID, startOfDay, endOfDay).Scan(&data.TodayCount)
	if err != nil { logger.Ctx(ctx).Error("Failed to get today count", logger.Error(err)); return data, err }

	err = s.pool.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM booking_appointments WHERE owner_id = ? AND start_time >= ? AND status IN ('pending','confirmed')",
		ownerID, endOfDay).Scan(&data.UpcomingCount)
	if err != nil { logger.Ctx(ctx).Error("Failed to get upcoming count", logger.Error(err)); return data, err }

	var totalMonth, cancelledMonth int
	err = s.pool.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM booking_appointments WHERE owner_id = ? AND start_time >= ?",
		ownerID, startOfMonth).Scan(&totalMonth)
	if err != nil { logger.Ctx(ctx).Error("Failed to get total month count", logger.Error(err)); return data, err }

	err = s.pool.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM booking_appointments WHERE owner_id = ? AND start_time >= ? AND status = 'cancelled'",
		ownerID, startOfMonth).Scan(&cancelledMonth)
	if err != nil { logger.Ctx(ctx).Error("Failed to get cancelled month count", logger.Error(err)); return data, err }

	if totalMonth > 0 {
		data.CancellationRate = float64(cancelledMonth) / float64(totalMonth) * 100
	}

	var revenue sql.NullFloat64
	err = s.pool.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(bs.price), 0) FROM booking_appointments ba
		 JOIN booking_services bs ON ba.service_id = bs.id
		 WHERE ba.owner_id = ? AND ba.start_time >= ? AND ba.status IN ('completed','confirmed')`,
		ownerID, startOfMonth).Scan(&revenue)
	if err != nil { logger.Ctx(ctx).Error("Failed to get revenue", logger.Error(err)); return data, err }
	if revenue.Valid { data.Revenue = revenue.Float64 }

	return data, nil
}
