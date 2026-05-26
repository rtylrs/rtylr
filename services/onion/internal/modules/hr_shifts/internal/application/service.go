package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/hr_shifts/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_shifts/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	hrShiftsScope = "hr_shifts"
	hrShiftsTTL   = 2 * time.Minute
)

func invalidateHRShiftsScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(hrShiftsScope, ownerID))
	}
}

type repository interface {
	ListShifts(ctx context.Context, ownerID int64, params domain.ListShiftsParams) ([]domain.ShiftRecord, int64, error)
	GetShift(ctx context.Context, ownerID, shiftID int64) (*domain.ShiftRecord, error)
	CreateShift(ctx context.Context, ownerID int64, shift domain.ShiftRecord) (*domain.ShiftRecord, error)
	UpdateShift(ctx context.Context, ownerID int64, shift domain.ShiftRecord) error
	ListShiftsByLocation(ctx context.Context, ownerID, locationID int64) ([]domain.ShiftRecord, error)
	ListRosterAssignments(ctx context.Context, ownerID int64, weekStart, weekEnd string) ([]domain.ShiftAssignmentRecord, error)
	CreateShiftAssignment(ctx context.Context, ownerID int64, assignment domain.ShiftAssignmentRecord) (*domain.ShiftAssignmentRecord, error)
	PublishRosterAssignments(ctx context.Context, ownerID int64, weekStart, weekEnd string) error
	ListEmployeeAssignmentsForDate(ctx context.Context, ownerID, employeeID int64, date string) ([]domain.ShiftAssignmentRecord, error)
}

type Service struct {
	repo                 repository
	leaveConflictChecker domain.LeaveConflictChecker
}

func NewService(_ context.Context, db *sql.DB, leaveChecker domain.LeaveConflictChecker) *Service {
	return &Service{
		repo:                 infrastructure.NewRepository(db),
		leaveConflictChecker: leaveChecker,
	}
}

func (s *Service) ListShifts(ctx context.Context, ownerID int64, params domain.ListShiftsParams) ([]domain.ShiftRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	type lr struct {
		Items []domain.ShiftRecord
		Total int64
	}
	key := cacheKeys.Scoped(hrShiftsScope, ownerID, fmt.Sprintf("list|l=%d|o=%d", params.Limit, params.Offset))
	res, err := cacher.Of[lr]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (lr, error) {
			items, total, err := s.repo.ListShifts(ctx, ownerID, params)
			return lr{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrShiftsTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) CreateShift(ctx context.Context, ownerID int64, shift domain.ShiftRecord) (*domain.ShiftRecord, error) {
	if shift.Color == "" {
		shift.Color = "#4CAF50"
	}
	rec, err := s.repo.CreateShift(ctx, ownerID, shift)
	if err == nil {
		invalidateHRShiftsScope(ctx, ownerID)
	}
	return rec, err
}

func (s *Service) UpdateShift(ctx context.Context, ownerID int64, shift domain.ShiftRecord) error {
	if err := s.repo.UpdateShift(ctx, ownerID, shift); err != nil {
		return err
	}
	invalidateHRShiftsScope(ctx, ownerID)
	return nil
}

func (s *Service) GetRoster(ctx context.Context, ownerID int64, params domain.RosterParams) ([]domain.ShiftAssignmentRecord, []domain.ConflictInfo, error) {
	weekStart, err := time.Parse("2006-01-02", params.WeekStart)
	if err != nil {
		return nil, nil, err
	}
	weekEnd := weekStart.AddDate(0, 0, 7).Format("2006-01-02")

	assignments, err := s.repo.ListRosterAssignments(ctx, ownerID, params.WeekStart, weekEnd)
	if err != nil {
		return nil, nil, err
	}

	var conflicts []domain.ConflictInfo
	seen := make(map[string]bool)
	for _, a := range assignments {
		key := fmt.Sprintf("%s-%d", a.Date, a.EmployeeID)

		if s.leaveConflictChecker != nil {
			hasLeave, leaveErr := s.leaveConflictChecker.HasApprovedLeave(ctx, ownerID, a.EmployeeID, a.Date)
			if leaveErr != nil {
				logger.Ctx(ctx).Error("failed to check leave conflict",
					logger.Any("owner_id", ownerID), logger.Any("employee_id", a.EmployeeID), logger.Error(leaveErr))
			}
			if hasLeave {
				conflicts = append(conflicts, domain.ConflictInfo{
					EmployeeID:   a.EmployeeID,
					Date:         a.Date,
					ConflictType: "leave",
					Description:  "Employee has approved leave on this date",
				})
			}
		}

		if seen[key] {
			conflicts = append(conflicts, domain.ConflictInfo{
				EmployeeID:   a.EmployeeID,
				Date:         a.Date,
				ConflictType: "double_shift",
				Description:  "Employee is assigned to multiple shifts on this date",
			})
		}
		seen[key] = true
	}

	return assignments, conflicts, nil
}

func (s *Service) CreateShiftAssignment(ctx context.Context, ownerID int64, assignment domain.ShiftAssignmentRecord) (*domain.ShiftAssignmentRecord, error) {
	_, err := s.repo.GetShift(ctx, ownerID, assignment.ShiftID)
	if err != nil {
		return nil, err
	}

	conflicts := s.DetectConflicts(ctx, ownerID, assignment.EmployeeID, assignment.Date)
	for _, c := range conflicts {
		if c.ConflictType == "leave" {
			return nil, domain.ErrRosterConflict
		}
	}

	rec, err := s.repo.CreateShiftAssignment(ctx, ownerID, assignment)
	if err == nil {
		invalidateHRShiftsScope(ctx, ownerID)
	}
	return rec, err
}

func (s *Service) PublishRoster(ctx context.Context, ownerID, locationID int64, weekStart string) error {
	ws, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		return err
	}
	weekEnd := ws.AddDate(0, 0, 7).Format("2006-01-02")

	if err := s.repo.PublishRosterAssignments(ctx, ownerID, weekStart, weekEnd); err != nil {
		return err
	}

	invalidateHRShiftsScope(ctx, ownerID)
	logger.Ctx(ctx).Info("roster published",
		logger.Any("owner_id", ownerID),
		logger.Any("location_id", locationID),
		logger.Any("week_start", weekStart))

	return nil
}

func (s *Service) DetectConflicts(ctx context.Context, ownerID, employeeID int64, date string) []domain.ConflictInfo {
	var conflicts []domain.ConflictInfo

	if s.leaveConflictChecker != nil {
		hasLeave, err := s.leaveConflictChecker.HasApprovedLeave(ctx, ownerID, employeeID, date)
		if err != nil {
			logger.Ctx(ctx).Error("failed to check leave for conflict detection",
				logger.Any("owner_id", ownerID), logger.Any("employee_id", employeeID), logger.Error(err))
		}
		if hasLeave {
			conflicts = append(conflicts, domain.ConflictInfo{
				EmployeeID:   employeeID,
				Date:         date,
				ConflictType: "leave",
				Description:  "Employee has approved leave on this date",
			})
		}
	}

	existing, err := s.repo.ListEmployeeAssignmentsForDate(ctx, ownerID, employeeID, date)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list existing assignments for conflict detection",
			logger.Any("owner_id", ownerID), logger.Any("employee_id", employeeID), logger.Error(err))
	}
	if len(existing) > 0 {
		conflicts = append(conflicts, domain.ConflictInfo{
			EmployeeID:   employeeID,
			Date:         date,
			ConflictType: "existing_shift",
			Description:  "Employee already has a shift assignment on this date",
		})
	}

	return conflicts
}
