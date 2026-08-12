package application

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/hr_shifts/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// ConflictTypeAvailability marks a conflict between a shift assignment and the
// assignee's declared availability. Unlike a "leave" conflict it never blocks
// the write; see CreateShiftAssignment.
const ConflictTypeAvailability = "availability"

// mondayFirstWeekday converts a date to the 0 = Monday .. 6 = Sunday numbering
// hr_staff_availability.day_of_week uses. Go's time.Weekday is Sunday = 0, so
// this is not a straight cast.
func mondayFirstWeekday(t time.Time) int32 {
	return int32((int(t.Weekday()) + 6) % 7)
}

// ListAvailability returns availability entries. A caller may always read their
// own; reading anyone else's, including an unfiltered listing that spans the
// whole tenant, requires manage_schedules.
func (s *Service) ListAvailability(ctx context.Context, businessID, callerUserID int64, params domain.ListAvailabilityParams) ([]domain.AvailabilityRecord, int64, error) {
	callerStaffID, err := s.callerStaffID(ctx, businessID, callerUserID)
	if err != nil && !errors.Is(err, domain.ErrStaffNotLinked) {
		return nil, 0, err
	}

	// Reading only your own rows never needs a permission.
	if params.StaffID == 0 || params.StaffID != callerStaffID {
		allowed, permErr := s.callerHasManageSchedules(ctx, businessID, callerUserID)
		if permErr != nil {
			return nil, 0, permErr
		}
		if !allowed {
			// Without the permission, silently narrow the listing to the caller's
			// own rows rather than refusing, so the self-service screen works for
			// everyone. A caller who explicitly asked for someone else is denied.
			if params.StaffID != 0 && params.StaffID != callerStaffID {
				return nil, 0, domain.ErrPermissionDenied
			}
			if callerStaffID == 0 {
				return nil, 0, domain.ErrStaffNotLinked
			}
			params.StaffID = callerStaffID
		}
	}

	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	return s.repo.ListAvailability(ctx, businessID, params)
}

// SetAvailability creates or updates one availability entry. Writing your own is
// always allowed; writing anyone else's requires manage_schedules.
func (s *Service) SetAvailability(ctx context.Context, businessID, callerUserID int64, entry domain.AvailabilityRecord) (*domain.AvailabilityRecord, error) {
	if err := domain.ValidateAvailability(entry); err != nil {
		return nil, err
	}

	if err := s.authorizeAvailabilityWrite(ctx, businessID, callerUserID, entry.StaffID); err != nil {
		return nil, err
	}

	rec, err := s.repo.UpsertAvailability(ctx, businessID, entry)
	if err != nil {
		return nil, err
	}
	invalidateHRShiftsScope(ctx, businessID)
	return rec, nil
}

// DeleteAvailability removes one entry. The entry is loaded first so the write
// is authorised against the staff member it actually belongs to, not against a
// staff id the caller supplied.
func (s *Service) DeleteAvailability(ctx context.Context, businessID, callerUserID, id int64) error {
	existing, err := s.repo.GetAvailability(ctx, businessID, id)
	if err != nil {
		return err
	}

	if err := s.authorizeAvailabilityWrite(ctx, businessID, callerUserID, existing.StaffID); err != nil {
		return err
	}

	if err := s.repo.DeleteAvailability(ctx, businessID, id); err != nil {
		return err
	}
	invalidateHRShiftsScope(ctx, businessID)
	return nil
}

// authorizeAvailabilityWrite permits a write to the caller's own entries and
// otherwise requires manage_schedules.
func (s *Service) authorizeAvailabilityWrite(ctx context.Context, businessID, callerUserID, targetStaffID int64) error {
	callerStaffID, err := s.callerStaffID(ctx, businessID, callerUserID)
	if err != nil && !errors.Is(err, domain.ErrStaffNotLinked) {
		return err
	}
	if callerStaffID != 0 && callerStaffID == targetStaffID {
		return nil
	}

	allowed, err := s.callerHasManageSchedules(ctx, businessID, callerUserID)
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrPermissionDenied
	}
	return nil
}

func (s *Service) callerStaffID(ctx context.Context, businessID, callerUserID int64) (int64, error) {
	if s.staffResolver == nil {
		return 0, domain.ErrStaffNotLinked
	}
	return s.staffResolver.StaffIDByUserID(ctx, businessID, callerUserID)
}

// callerHasManageSchedules resolves the caller to an employee and checks the
// permission, reusing the same provider hr_payroll gates on.
func (s *Service) callerHasManageSchedules(ctx context.Context, businessID, callerUserID int64) (bool, error) {
	if s.permissions == nil {
		return false, nil
	}
	employeeID, err := s.permissions.GetEmployeeIDByUserID(ctx, businessID, callerUserID)
	if err != nil {
		// No employee record means no role, so no permission. That is a denial,
		// not a failure of the request.
		logger.Ctx(ctx).Debug("no employee record for caller during permission check",
			logger.Any("business_id", businessID), logger.Any("user_id", callerUserID), logger.Error(err))
		return false, nil
	}
	return s.permissions.HasPermission(ctx, businessID, employeeID, domain.PermissionManageSchedules)
}

// detectAvailabilityConflicts reports whether assigning employeeID a shift
// running start..end on date contradicts their declared availability.
//
// Precedence follows the contract: if any dated exception exists for that date
// it governs the day outright and the recurring pattern is ignored. A staff
// member who has declared nothing gets no warning, because silence is not a
// refusal.
func (s *Service) detectAvailabilityConflicts(ctx context.Context, businessID, employeeID int64, date, shiftStart, shiftEnd string) []domain.ConflictInfo {
	if s.staffResolver == nil {
		return nil
	}

	staffID, err := s.staffResolver.StaffIDByEmployeeID(ctx, businessID, employeeID)
	if err != nil {
		if errors.Is(err, domain.ErrStaffNotLinked) {
			// The employee has no hr_staff row, so there is no availability to
			// read. Logged because it also means this check is silently inert
			// for that employee.
			logger.Ctx(ctx).Debug("skipping availability check, employee has no linked staff record",
				logger.Any("business_id", businessID), logger.Any("employee_id", employeeID))
			return nil
		}
		logger.Ctx(ctx).Error("failed to resolve staff record for availability check",
			logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
		return nil
	}

	day, err := time.Parse(time.DateOnly, date)
	if err != nil {
		logger.Ctx(ctx).Error("failed to parse assignment date for availability check",
			logger.Any("business_id", businessID), logger.Any("date", date), logger.Error(err))
		return nil
	}

	entries, err := s.repo.ListAvailabilityForStaffDate(ctx, businessID, staffID, date, mondayFirstWeekday(day))
	if err != nil {
		logger.Ctx(ctx).Error("failed to read availability for conflict detection",
			logger.Any("business_id", businessID), logger.Any("staff_id", staffID), logger.Error(err))
		return nil
	}
	if len(entries) == 0 {
		return nil
	}

	governing := governingEntries(entries)
	shift, ok := parseWindow(shiftStart, shiftEnd)
	if !ok {
		return nil
	}

	var blocked []domain.AvailabilityRecord
	var offered []domain.AvailabilityRecord
	for _, e := range governing {
		if e.IsAvailable {
			offered = append(offered, e)
			continue
		}
		blocked = append(blocked, e)
	}

	for _, b := range blocked {
		w, valid := parseWindow(b.StartTime, b.EndTime)
		if valid && windowsOverlap(shift, w) {
			return []domain.ConflictInfo{{
				EmployeeID:   employeeID,
				Date:         date,
				ConflictType: ConflictTypeAvailability,
				Description: fmt.Sprintf("Staff member is marked unavailable %s-%s on %s",
					b.StartTime, b.EndTime, date),
			}}
		}
	}

	if len(offered) == 0 {
		return nil
	}
	for _, o := range offered {
		w, valid := parseWindow(o.StartTime, o.EndTime)
		if valid && windowContains(w, shift) {
			return nil
		}
	}

	return []domain.ConflictInfo{{
		EmployeeID:   employeeID,
		Date:         date,
		ConflictType: ConflictTypeAvailability,
		Description: fmt.Sprintf("Shift %s-%s on %s falls outside the staff member's declared availability",
			shiftStart, shiftEnd, date),
	}}
}

// governingEntries applies the contract's precedence rule: a dated exception
// replaces, rather than adds to, the recurring rules for that date.
func governingEntries(entries []domain.AvailabilityRecord) []domain.AvailabilityRecord {
	var exceptions []domain.AvailabilityRecord
	for _, e := range entries {
		if e.Kind() == domain.AvailabilityKindException {
			exceptions = append(exceptions, e)
		}
	}
	if len(exceptions) > 0 {
		return exceptions
	}
	return entries
}

// segment is a half-open minute range within a single day.
type segment struct{ from, to int }

// parseWindow turns an HH:MM pair into one or two segments. An end earlier than
// the start means the window crosses midnight, which the contract uses to
// express a night shift, so it becomes two segments rather than an empty range.
func parseWindow(start, end string) ([]segment, bool) {
	s, okStart := minutesOfDay(start)
	e, okEnd := minutesOfDay(end)
	if !okStart || !okEnd {
		return nil, false
	}
	if s == e {
		return nil, false
	}
	if s < e {
		return []segment{{from: s, to: e}}, true
	}
	return []segment{{from: s, to: 24 * 60}, {from: 0, to: e}}, true
}

func minutesOfDay(v string) (int, bool) {
	layouts := []string{"15:04", "15:04:05"}
	for _, layout := range layouts {
		t, err := time.Parse(layout, v)
		if err == nil {
			return t.Hour()*60 + t.Minute(), true
		}
	}
	return 0, false
}

func windowsOverlap(a, b []segment) bool {
	for _, x := range a {
		for _, y := range b {
			if x.from < y.to && y.from < x.to {
				return true
			}
		}
	}
	return false
}

// windowContains reports whether every part of inner falls inside outer.
func windowContains(outer, inner []segment) bool {
	for _, i := range inner {
		covered := false
		for _, o := range outer {
			if o.from <= i.from && i.to <= o.to {
				covered = true
				break
			}
		}
		if !covered {
			return false
		}
	}
	return true
}

// describeConflicts renders conflicts as the human-readable strings the contract
// carries on CreateShiftAssignmentResponse.warnings.
func describeConflicts(conflicts []domain.ConflictInfo) []string {
	if len(conflicts) == 0 {
		return nil
	}
	out := make([]string, 0, len(conflicts))
	for _, c := range conflicts {
		out = append(out, c.Description+" (employee "+strconv.FormatInt(c.EmployeeID, 10)+")")
	}
	return out
}
