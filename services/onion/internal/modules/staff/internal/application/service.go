package application

import (
	"context"
	"database/sql"
	"strconv"
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

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func (s *Service) ListStaff(ctx context.Context, businessID int64, search, departmentID, status string) ([]domain.Staff, error) {
	return s.repo.ListStaff(ctx, businessID, search, departmentID, status)
}

// ListStaffDirectory returns one page of the business-wide staff roster plus
// the total count, for the directory endpoint's permission-gated caller.
func (s *Service) ListStaffDirectory(ctx context.Context, businessID, limit, offset int64) ([]domain.Staff, int64, error) {
	staff, err := s.repo.ListStaffDirectory(ctx, businessID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountStaffDirectory(ctx, businessID)
	if err != nil {
		return nil, 0, err
	}
	return staff, total, nil
}

func (s *Service) CreateStaff(ctx context.Context, input domain.StaffInput) (int64, error) {
	return s.repo.CreateStaff(ctx, input)
}

func (s *Service) UpdateStaff(ctx context.Context, businessID int64, input domain.StaffUpdateInput) error {
	return s.repo.UpdateStaff(ctx, businessID, input)
}

func (s *Service) GetStaffByID(ctx context.Context, businessID, staffID int64) (*domain.Staff, error) {
	return s.repo.GetStaffByID(ctx, businessID, staffID)
}

func (s *Service) DeleteStaff(ctx context.Context, businessID, staffID int64) error {
	return s.repo.DeleteStaff(ctx, businessID, staffID)
}

func (s *Service) ClockIn(ctx context.Context, businessID, staffID int64) error {
	return s.repo.ClockIn(ctx, businessID, staffID)
}

func (s *Service) ClockOut(ctx context.Context, businessID, staffID int64) error {
	return s.repo.ClockOut(ctx, businessID, staffID)
}

func (s *Service) GetDashboardStats(ctx context.Context, businessID int64) (domain.DashboardStats, error) {
	cacheKey := "staff:dashboard:" + strconv.FormatInt(businessID, 10)

	return cacher.Of[domain.DashboardStats]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) (domain.DashboardStats, error) {
		var stats domain.DashboardStats

		total, err := s.repo.CountActiveStaff(ctx, businessID)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to count active staff", logger.Error(err))
			return stats, err
		}

		onShift, err := s.repo.CountOnShiftToday(ctx, businessID)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to count on-shift today", logger.Error(err))
			return stats, err
		}

		pendingTimeOff, err := s.repo.CountPendingTimeOff(ctx, businessID)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to count pending time off", logger.Error(err))
			return stats, err
		}

		payrollDraft, err := s.repo.CountDraftPayroll(ctx, businessID)
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

// ListManagedUserIDs returns the user ids reporting to managerUserID. It backs
// the orgchart provider Flow uses for the dept-manager visibility proxy, and
// deliberately exposes ids only: no staff record, and nothing on it, crosses
// the module boundary.
func (s *Service) ListManagedUserIDs(ctx context.Context, businessID, managerUserID int64) ([]int64, error) {
	if businessID <= 0 || managerUserID <= 0 {
		return nil, nil
	}
	return s.repo.ListManagedUserIDs(ctx, businessID, managerUserID)
}
