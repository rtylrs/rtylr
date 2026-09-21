package application

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
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

func invalidateHRShiftsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(hrShiftsScope, businessID))
}

type repository interface {
	ListShifts(ctx context.Context, businessID int64, params domain.ListShiftsParams) ([]domain.ShiftRecord, int64, error)
	GetShift(ctx context.Context, businessID, shiftID int64) (*domain.ShiftRecord, error)
	CreateShift(ctx context.Context, businessID int64, shift domain.ShiftRecord) (*domain.ShiftRecord, error)
	UpdateShift(ctx context.Context, businessID int64, shift domain.ShiftRecord) error
	DeleteShift(ctx context.Context, businessID, shiftID int64) error
	ListShiftsByLocation(ctx context.Context, businessID, locationID int64) ([]domain.ShiftRecord, error)
	ListRosterAssignments(ctx context.Context, businessID int64, weekStart, weekEnd string) ([]domain.ShiftAssignmentRecord, error)
	CreateShiftAssignment(ctx context.Context, businessID int64, assignment domain.ShiftAssignmentRecord) (*domain.ShiftAssignmentRecord, error)
	PublishRosterAssignments(ctx context.Context, businessID int64, weekStart, weekEnd string) error
	ListEmployeeAssignmentsForDate(ctx context.Context, businessID, employeeID int64, date string) ([]domain.ShiftAssignmentRecord, error)

	ListAvailability(ctx context.Context, businessID int64, params domain.ListAvailabilityParams) ([]domain.AvailabilityRecord, int64, error)
	GetAvailability(ctx context.Context, businessID, id int64) (*domain.AvailabilityRecord, error)
	UpsertAvailability(ctx context.Context, businessID int64, entry domain.AvailabilityRecord) (*domain.AvailabilityRecord, error)
	DeleteAvailability(ctx context.Context, businessID, id int64) error
	ListAvailabilityForStaffDate(ctx context.Context, businessID, staffID int64, date string, dayOfWeek int32) ([]domain.AvailabilityRecord, error)
}

type Service struct {
	repo                 repository
	leaveConflictChecker domain.LeaveConflictChecker
	// permissions gates writes against another staff member's availability. Nil
	// means no provider was wired, which is treated as "no permission".
	permissions domain.PermissionProvider
	// staffResolver maps between hr_staff.id and hr_employees.id. Nil disables
	// the availability check rather than failing the assignment.
	staffResolver domain.StaffResolver
}

func NewService(
	_ context.Context,
	db *sql.DB,
	leaveChecker domain.LeaveConflictChecker,
	permissions domain.PermissionProvider,
) *Service {
	repo := infrastructure.NewRepository(db)
	return &Service{
		repo:                 repo,
		leaveConflictChecker: leaveChecker,
		permissions:          permissions,
		staffResolver:        repo,
	}
}

func (s *Service) ListShifts(ctx context.Context, businessID int64, params domain.ListShiftsParams) ([]domain.ShiftRecord, int64, error) {
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
	// Every filter has to be part of the key, or one filter's result set gets
	// served to a request using a different one.
	key := cacheKeys.Scoped(hrShiftsScope, businessID, fmt.Sprintf(
		"list|l=%d|o=%d|staff=%d|date=%s|loc=%d|kind=%s",
		params.Limit, params.Offset, params.StaffID, params.Date, params.LocationID, params.Kind,
	))
	res, err := cacher.Of[lr]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (lr, error) {
			items, total, err := s.repo.ListShifts(ctx, businessID, params)
			return lr{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrShiftsTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) CreateShift(ctx context.Context, businessID int64, shift domain.ShiftRecord) (*domain.ShiftRecord, error) {
	if shift.Color == "" {
		shift.Color = domain.DefaultShiftColor
	}
	if shift.Status == "" {
		shift.Status = domain.DefaultShiftStatus
	}
	if err := domain.ValidateShift(shift); err != nil {
		return nil, err
	}

	rec, err := s.repo.CreateShift(ctx, businessID, shift)
	if err == nil {
		invalidateHRShiftsScope(ctx, businessID)
	}
	return rec, err
}

// UpdateShift applies a partial edit. Fields absent from the request keep their
// stored value, so a caller editing one field cannot blank the others.
func (s *Service) UpdateShift(ctx context.Context, businessID int64, shift domain.ShiftRecord) (*domain.ShiftRecord, error) {
	existing, err := s.repo.GetShift(ctx, businessID, shift.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrShiftNotFound
	}

	merged := mergeShift(*existing, shift)
	if err := domain.ValidateShift(merged); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateShift(ctx, businessID, merged); err != nil {
		return nil, err
	}
	invalidateHRShiftsScope(ctx, businessID)

	return s.repo.GetShift(ctx, businessID, shift.ID)
}

func (s *Service) DeleteShift(ctx context.Context, businessID, shiftID int64) error {
	if err := s.repo.DeleteShift(ctx, businessID, shiftID); err != nil {
		return err
	}
	invalidateHRShiftsScope(ctx, businessID)
	return nil
}

// mergeShift overlays the populated fields of patch onto existing. StaffID is
// deliberately not overlaid: it is the discriminator between a template and an
// attendance row, so an update must not silently convert one kind into the other.
func mergeShift(existing, patch domain.ShiftRecord) domain.ShiftRecord {
	merged := existing
	merged.ID = existing.ID

	if patch.LocationID != 0 {
		merged.LocationID = patch.LocationID
	}
	if patch.Name != "" {
		merged.Name = patch.Name
	}
	if patch.StartTime != "" {
		merged.StartTime = patch.StartTime
	}
	if patch.EndTime != "" {
		merged.EndTime = patch.EndTime
	}
	if patch.Color != "" {
		merged.Color = patch.Color
	}
	if patch.DepartmentID != 0 {
		merged.DepartmentID = patch.DepartmentID
	}
	if patch.Date != "" {
		merged.Date = patch.Date
	}
	if patch.BreakDurationMinutes != 0 {
		merged.BreakDurationMinutes = patch.BreakDurationMinutes
	}
	if patch.Status != "" {
		merged.Status = patch.Status
	}
	if patch.Notes != "" {
		merged.Notes = patch.Notes
	}

	return merged
}

func (s *Service) GetRoster(ctx context.Context, businessID int64, params domain.RosterParams) ([]domain.ShiftAssignmentRecord, []domain.ConflictInfo, error) {
	weekStart, err := time.Parse("2006-01-02", params.WeekStart)
	if err != nil {
		return nil, nil, err
	}
	weekEnd := weekStart.AddDate(0, 0, 7).Format("2006-01-02")

	assignments, err := s.repo.ListRosterAssignments(ctx, businessID, params.WeekStart, weekEnd)
	if err != nil {
		return nil, nil, err
	}

	var conflicts []domain.ConflictInfo
	seen := make(map[string]bool)
	for _, a := range assignments {
		key := a.Date + "-" + strconv.FormatInt(a.EmployeeID, 10)

		if s.leaveConflictChecker != nil {
			hasLeave, leaveErr := s.leaveConflictChecker.HasApprovedLeave(ctx, businessID, a.EmployeeID, a.Date)
			if leaveErr != nil {
				logger.Ctx(ctx).Error("failed to check leave conflict",
					logger.Any("business_id", businessID), logger.Any("employee_id", a.EmployeeID), logger.Error(leaveErr))
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

// CreateShiftAssignment assigns an employee to a shift and returns any advisory
// warnings alongside the new row.
//
// Approved leave still blocks the write, because leave is an entitlement the
// business has already granted. Declared availability does not: it is a stated
// preference, and a scheduler legitimately needs to staff against one to cover a
// no-show or an agreed swap. Those conflicts come back as warnings so the caller
// is told without being stopped.
func (s *Service) CreateShiftAssignment(ctx context.Context, businessID int64, assignment domain.ShiftAssignmentRecord) (*domain.ShiftAssignmentRecord, []string, error) {
	shift, err := s.repo.GetShift(ctx, businessID, assignment.ShiftID)
	if err != nil {
		return nil, nil, err
	}

	conflicts := s.DetectConflicts(ctx, businessID, assignment.EmployeeID, assignment.Date)
	for _, c := range conflicts {
		// An existing assignment on the same date blocks too: uk_shift_assignment
		// allows only one assignment per (owner, employee, date), so letting the
		// insert through would surface a raw duplicate-key error as a 500 while
		// the endpoint's contract promises 409 for "another shift".
		if c.ConflictType == "leave" || c.ConflictType == "existing_shift" {
			return nil, nil, domain.ErrRosterConflict
		}
	}

	availabilityConflicts := s.detectAvailabilityConflicts(
		ctx, businessID, assignment.EmployeeID, assignment.Date, shift.StartTime, shift.EndTime,
	)

	rec, err := s.repo.CreateShiftAssignment(ctx, businessID, assignment)
	if err != nil {
		return nil, nil, err
	}
	invalidateHRShiftsScope(ctx, businessID)

	warnings := describeConflicts(availabilityConflicts)
	if len(warnings) > 0 {
		logger.Ctx(ctx).Info("shift assignment created against declared availability",
			logger.Any("business_id", businessID),
			logger.Any("employee_id", assignment.EmployeeID),
			logger.Any("date", assignment.Date),
			logger.Any("warning_count", len(warnings)))
	}

	return rec, warnings, nil
}

func (s *Service) PublishRoster(ctx context.Context, businessID, locationID int64, weekStart string) error {
	ws, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		return err
	}
	weekEnd := ws.AddDate(0, 0, 7).Format("2006-01-02")

	if err := s.repo.PublishRosterAssignments(ctx, businessID, weekStart, weekEnd); err != nil {
		return err
	}

	invalidateHRShiftsScope(ctx, businessID)
	logger.Ctx(ctx).Info("roster published",
		logger.Any("business_id", businessID),
		logger.Any("location_id", locationID),
		logger.Any("week_start", weekStart))

	return nil
}

func (s *Service) DetectConflicts(ctx context.Context, businessID, employeeID int64, date string) []domain.ConflictInfo {
	var conflicts []domain.ConflictInfo

	if s.leaveConflictChecker != nil {
		hasLeave, err := s.leaveConflictChecker.HasApprovedLeave(ctx, businessID, employeeID, date)
		if err != nil {
			logger.Ctx(ctx).Error("failed to check leave for conflict detection",
				logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
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

	existing, err := s.repo.ListEmployeeAssignmentsForDate(ctx, businessID, employeeID, date)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list existing assignments for conflict detection",
			logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
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
