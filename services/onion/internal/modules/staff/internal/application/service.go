package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/staff/internal/domain"
	"github.com/voxire/onion/internal/modules/staff/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/logger"
)

const dashboardCacheTTL = 2 * time.Minute

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) ListStaff(ctx context.Context, ownerID int64, search, departmentID, status string) ([]domain.Staff, error) {
	return s.repo.ListStaff(ctx, ownerID, search, departmentID, status)
}

func (s *Service) CreateStaff(ctx context.Context, input domain.StaffInput) (int64, error) {
	return s.repo.CreateStaff(ctx, input)
}

func (s *Service) UpdateStaff(ctx context.Context, ownerID int64, input domain.StaffUpdateInput) error {
	return s.repo.UpdateStaff(ctx, ownerID, input)
}

func (s *Service) DeleteStaff(ctx context.Context, ownerID, staffID int64) error {
	return s.repo.DeleteStaff(ctx, ownerID, staffID)
}

func (s *Service) GetDashboardStats(ctx context.Context, ownerID int64) (domain.DashboardStats, error) {
	cacheKey := fmt.Sprintf("staff:dashboard:%d", ownerID)

	return cacher.Of[domain.DashboardStats]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) (domain.DashboardStats, error) {
		var stats domain.DashboardStats

		total, err := s.repo.CountActiveStaff(ctx, ownerID)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to count active staff", logger.Error(err))
			return stats, err
		}

		onShift, err := s.repo.CountOnShiftToday(ctx, ownerID)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to count on-shift today", logger.Error(err))
			return stats, err
		}

		pendingTimeOff, err := s.repo.CountPendingTimeOff(ctx, ownerID)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to count pending time off", logger.Error(err))
			return stats, err
		}

		payrollDraft, err := s.repo.CountDraftPayroll(ctx, ownerID)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to count draft payroll", logger.Error(err))
			return stats, err
		}

		stats.TotalStaff = total
		stats.OnShiftToday = onShift
		stats.PendingTimeOff = pendingTimeOff
		stats.PayrollDraft = payrollDraft

		return stats, nil
	}, cacher.WithExpiry(ctx, dashboardCacheTTL))
}
