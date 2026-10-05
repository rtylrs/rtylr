package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/domain"
)

// The caller in these tests is user 900, who is staff 10. Staff 20 is somebody
// else, so a write aimed at 20 is the cross-staff case that needs the permission.
const (
	callerUserID  int64 = 900
	callerStaffID int64 = 10
	otherStaffID  int64 = 20
	testBusiness  int64 = 1001
)

// selfServiceService builds a service where the caller resolves to callerStaffID
// and holds no permissions at all. Anything it is allowed to do, it is allowed
// to do purely because the row is its own.
func selfServiceService(repo *mockRepo) *Service {
	return &Service{
		repo:          repo,
		staffResolver: &mockStaffResolver{staffByUser: callerStaffID},
		permissions:   &mockPermissions{employeeID: 5000, allowed: false},
	}
}

func managerService(repo *mockRepo) *Service {
	return &Service{
		repo:          repo,
		staffResolver: &mockStaffResolver{staffByUser: callerStaffID},
		permissions:   &mockPermissions{employeeID: 5000, allowed: true},
	}
}

func recurringEntry(staffID int64, day int32, start, end string, available bool) domain.AvailabilityRecord {
	return domain.AvailabilityRecord{
		StaffID:     staffID,
		DayOfWeek:   day,
		StartTime:   start,
		EndTime:     end,
		IsAvailable: available,
	}
}

func exceptionEntry(staffID int64, date, start, end string, available bool) domain.AvailabilityRecord {
	return domain.AvailabilityRecord{
		StaffID:      staffID,
		DayOfWeek:    domain.NoDayOfWeek,
		SpecificDate: date,
		StartTime:    start,
		EndTime:      end,
		IsAvailable:  available,
	}
}

// ---- self-service read and write of your own availability ----

func TestSetAvailability_OwnEntryAllowedWithoutPermission(t *testing.T) {
	repo := &mockRepo{}
	s := selfServiceService(repo)

	entry := recurringEntry(callerStaffID, 0, "09:00", "17:00", true)
	saved, err := s.SetAvailability(context.Background(), testBusiness, callerUserID, entry)
	if err != nil {
		t.Fatalf("writing own availability must not require a permission, got: %v", err)
	}
	if saved.StaffID != callerStaffID {
		t.Errorf("expected entry for staff %d, got %d", callerStaffID, saved.StaffID)
	}
	if repo.upserted.StartTime != "09:00" {
		t.Errorf("expected the entry to reach the repository, got %+v", repo.upserted)
	}
}

func TestListAvailability_OwnEntriesAllowedWithoutPermission(t *testing.T) {
	repo := &mockRepo{
		availability:      []domain.AvailabilityRecord{recurringEntry(callerStaffID, 0, "09:00", "17:00", true)},
		availabilityTotal: 1,
	}
	s := selfServiceService(repo)

	got, total, err := s.ListAvailability(context.Background(), testBusiness, callerUserID,
		domain.ListAvailabilityParams{StaffID: callerStaffID})
	if err != nil {
		t.Fatalf("reading own availability must not require a permission, got: %v", err)
	}
	if len(got) != 1 || total != 1 {
		t.Fatalf("expected 1 entry, got %d (total %d)", len(got), total)
	}
	if repo.lastAvailParams.StaffID != callerStaffID {
		t.Errorf("expected the query scoped to staff %d, got %d", callerStaffID, repo.lastAvailParams.StaffID)
	}
}

func TestDeleteAvailability_OwnEntryAllowedWithoutPermission(t *testing.T) {
	own := recurringEntry(callerStaffID, 0, "09:00", "17:00", true)
	own.ID = 77
	repo := &mockRepo{availabilityOne: &own}
	s := selfServiceService(repo)

	if err := s.DeleteAvailability(context.Background(), testBusiness, callerUserID, 77); err != nil {
		t.Fatalf("deleting own availability must not require a permission, got: %v", err)
	}
	if repo.deletedAvailID != 77 {
		t.Errorf("expected entry 77 deleted, got %d", repo.deletedAvailID)
	}
}

// ---- denial of writing someone else's without manage_schedules ----

func TestSetAvailability_OtherStaffDeniedWithoutManageSchedules(t *testing.T) {
	repo := &mockRepo{}
	s := selfServiceService(repo)

	entry := recurringEntry(otherStaffID, 0, "09:00", "17:00", true)
	_, err := s.SetAvailability(context.Background(), testBusiness, callerUserID, entry)
	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Fatalf("expected ErrPermissionDenied writing another staff member, got: %v", err)
	}
	if repo.upserted.StaffID != 0 {
		t.Error("denied write must not reach the repository")
	}
}

func TestSetAvailability_OtherStaffAllowedWithManageSchedules(t *testing.T) {
	repo := &mockRepo{}
	s := managerService(repo)

	entry := recurringEntry(otherStaffID, 0, "09:00", "17:00", true)
	saved, err := s.SetAvailability(context.Background(), testBusiness, callerUserID, entry)
	if err != nil {
		t.Fatalf("manage_schedules must permit writing another staff member, got: %v", err)
	}
	if saved.StaffID != otherStaffID {
		t.Errorf("expected entry for staff %d, got %d", otherStaffID, saved.StaffID)
	}
}

// DeleteAvailability authorises against the stored row's owner, not against a
// staff id the caller passes, so it cannot be steered by a forged request.
func TestDeleteAvailability_OtherStaffDeniedWithoutManageSchedules(t *testing.T) {
	theirs := recurringEntry(otherStaffID, 0, "09:00", "17:00", true)
	theirs.ID = 88
	repo := &mockRepo{availabilityOne: &theirs}
	s := selfServiceService(repo)

	err := s.DeleteAvailability(context.Background(), testBusiness, callerUserID, 88)
	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Fatalf("expected ErrPermissionDenied deleting another staff member's entry, got: %v", err)
	}
	if repo.availDeleteCalls != 0 {
		t.Error("denied delete must not reach the repository")
	}
}

func TestListAvailability_OtherStaffDeniedWithoutManageSchedules(t *testing.T) {
	repo := &mockRepo{}
	s := selfServiceService(repo)

	_, _, err := s.ListAvailability(context.Background(), testBusiness, callerUserID,
		domain.ListAvailabilityParams{StaffID: otherStaffID})
	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Fatalf("expected ErrPermissionDenied reading another staff member, got: %v", err)
	}
}

// An unfiltered listing spans the tenant, so without the permission it is
// narrowed to the caller rather than refused: the self-service screen sends no
// staff_id and must still work.
func TestListAvailability_UnfilteredNarrowsToSelfWithoutPermission(t *testing.T) {
	repo := &mockRepo{}
	s := selfServiceService(repo)

	if _, _, err := s.ListAvailability(context.Background(), testBusiness, callerUserID,
		domain.ListAvailabilityParams{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastAvailParams.StaffID != callerStaffID {
		t.Errorf("expected listing narrowed to staff %d, got %d", callerStaffID, repo.lastAvailParams.StaffID)
	}
}

func TestListAvailability_UnfilteredStaysTenantWideWithManageSchedules(t *testing.T) {
	repo := &mockRepo{}
	s := managerService(repo)

	if _, _, err := s.ListAvailability(context.Background(), testBusiness, callerUserID,
		domain.ListAvailabilityParams{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastAvailParams.StaffID != 0 {
		t.Errorf("expected an unfiltered listing, got staff filter %d", repo.lastAvailParams.StaffID)
	}
}

// ---- validation ----

func TestSetAvailability_RejectsBothWeekdayAndDate(t *testing.T) {
	s := selfServiceService(&mockRepo{})
	entry := recurringEntry(callerStaffID, 0, "09:00", "17:00", true)
	entry.SpecificDate = "2026-08-17"

	_, err := s.SetAvailability(context.Background(), testBusiness, callerUserID, entry)
	if !errors.Is(err, domain.ErrAvailabilityKindAmbiguous) {
		t.Fatalf("expected ErrAvailabilityKindAmbiguous, got: %v", err)
	}
}

func TestSetAvailability_RejectsNeitherWeekdayNorDate(t *testing.T) {
	s := selfServiceService(&mockRepo{})
	entry := domain.AvailabilityRecord{
		StaffID:   callerStaffID,
		DayOfWeek: domain.NoDayOfWeek,
		StartTime: "09:00",
		EndTime:   "17:00",
	}

	_, err := s.SetAvailability(context.Background(), testBusiness, callerUserID, entry)
	if !errors.Is(err, domain.ErrAvailabilityKindAmbiguous) {
		t.Fatalf("expected ErrAvailabilityKindAmbiguous, got: %v", err)
	}
}

func TestSetAvailability_RejectsWeekdayOutOfRange(t *testing.T) {
	s := selfServiceService(&mockRepo{})
	_, err := s.SetAvailability(context.Background(), testBusiness, callerUserID,
		recurringEntry(callerStaffID, 7, "09:00", "17:00", true))
	if !errors.Is(err, domain.ErrAvailabilityDayOutOfRange) {
		t.Fatalf("expected ErrAvailabilityDayOutOfRange, got: %v", err)
	}
}

// ---- the assignment-time warning ----

// assignmentService wires a service whose shift runs 09:00-17:00 and whose
// employee 5 resolves to staff 10.
func assignmentService(forDate []domain.AvailabilityRecord) (*Service, *mockRepo) {
	repo := &mockRepo{
		shift:   &domain.ShiftRecord{ID: 1, StartTime: "09:00:00", EndTime: "17:00:00"},
		forDate: forDate,
	}
	s := &Service{
		repo:          repo,
		staffResolver: &mockStaffResolver{staffByUser: callerStaffID, staffByEmployee: callerStaffID},
	}
	return s, repo
}

func assign(t *testing.T, s *Service, date string) []string {
	t.Helper()
	_, warnings, err := s.CreateShiftAssignment(context.Background(), testBusiness,
		domain.ShiftAssignmentRecord{ShiftID: 1, EmployeeID: 5, Date: date})
	if err != nil {
		t.Fatalf("assignment must succeed even when it conflicts: %v", err)
	}
	return warnings
}

// 2026-08-17 is a Monday, which is day_of_week 0 under the contract's
// Monday-first numbering.
const mondayDate = "2026-08-17"

func TestCreateShiftAssignment_WarnsWhenMarkedUnavailable(t *testing.T) {
	s, _ := assignmentService([]domain.AvailabilityRecord{
		exceptionEntry(callerStaffID, mondayDate, "09:00", "12:00", false),
	})

	warnings := assign(t, s, mondayDate)
	if len(warnings) != 1 {
		t.Fatalf("expected exactly 1 warning, got %d: %v", len(warnings), warnings)
	}
}

func TestCreateShiftAssignment_WarnsWhenOutsideDeclaredAvailability(t *testing.T) {
	s, _ := assignmentService([]domain.AvailabilityRecord{
		recurringEntry(callerStaffID, 0, "13:00", "18:00", true),
	})

	warnings := assign(t, s, mondayDate)
	if len(warnings) != 1 {
		t.Fatalf("expected a warning for a shift outside the declared window, got %v", warnings)
	}
}

func TestCreateShiftAssignment_SilentWhenShiftInsideAvailability(t *testing.T) {
	s, _ := assignmentService([]domain.AvailabilityRecord{
		recurringEntry(callerStaffID, 0, "08:00", "18:00", true),
	})

	if warnings := assign(t, s, mondayDate); len(warnings) != 0 {
		t.Fatalf("expected no warning when the shift fits the declared window, got %v", warnings)
	}
}

// Silence is not a refusal: a staff member who has declared nothing is not
// warned about, or every unconfigured tenant would warn on every assignment.
func TestCreateShiftAssignment_SilentWhenNothingDeclared(t *testing.T) {
	s, _ := assignmentService(nil)

	if warnings := assign(t, s, mondayDate); len(warnings) != 0 {
		t.Fatalf("expected no warning when no availability is declared, got %v", warnings)
	}
}

// A dated exception replaces the recurring pattern rather than adding to it, so
// an exception that grants availability silences a recurring block.
func TestCreateShiftAssignment_ExceptionOverridesRecurringBlock(t *testing.T) {
	s, _ := assignmentService([]domain.AvailabilityRecord{
		exceptionEntry(callerStaffID, mondayDate, "08:00", "18:00", true),
		recurringEntry(callerStaffID, 0, "09:00", "12:00", false),
	})

	if warnings := assign(t, s, mondayDate); len(warnings) != 0 {
		t.Fatalf("a dated exception must override the recurring pattern, got %v", warnings)
	}
}

func TestCreateShiftAssignment_ExceptionOverridesRecurringGrant(t *testing.T) {
	s, _ := assignmentService([]domain.AvailabilityRecord{
		exceptionEntry(callerStaffID, mondayDate, "09:00", "12:00", false),
		recurringEntry(callerStaffID, 0, "08:00", "18:00", true),
	})

	if warnings := assign(t, s, mondayDate); len(warnings) != 1 {
		t.Fatalf("a dated block must override the recurring grant, got %v", warnings)
	}
}

// The weekday handed to the repository must be Monday-first, not Go's
// Sunday-first time.Weekday.
func TestCreateShiftAssignment_UsesMondayFirstWeekday(t *testing.T) {
	s, repo := assignmentService(nil)

	assign(t, s, mondayDate)
	if repo.forDateDayOfWeek != 0 {
		t.Errorf("expected Monday to be day 0, got %d", repo.forDateDayOfWeek)
	}

	// 2026-08-20 is the Thursday of the same week.
	assign(t, s, "2026-08-20")
	if repo.forDateDayOfWeek != 3 {
		t.Errorf("expected Thursday to be day 3, got %d", repo.forDateDayOfWeek)
	}
}

// An employee with no linked hr_staff row has no availability to read. The
// assignment still succeeds and the check is simply inert.
func TestCreateShiftAssignment_SilentWhenStaffNotLinked(t *testing.T) {
	repo := &mockRepo{shift: &domain.ShiftRecord{ID: 1, StartTime: "09:00:00", EndTime: "17:00:00"}}
	s := &Service{
		repo:          repo,
		staffResolver: &mockStaffResolver{employeeErr: domain.ErrStaffNotLinked},
	}

	_, warnings, err := s.CreateShiftAssignment(context.Background(), testBusiness,
		domain.ShiftAssignmentRecord{ShiftID: 1, EmployeeID: 5, Date: mondayDate})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warning for an unlinked employee, got %v", warnings)
	}
	if repo.forDateCalls != 0 {
		t.Error("availability must not be queried when no staff row is linked")
	}
}

// A night shift is expressed with end_time before start_time. The overlap test
// must treat that as a window crossing midnight, not as an empty range.
func TestCreateShiftAssignment_WarnsAcrossMidnight(t *testing.T) {
	repo := &mockRepo{
		shift: &domain.ShiftRecord{ID: 1, StartTime: "22:00:00", EndTime: "06:00:00"},
		forDate: []domain.AvailabilityRecord{
			exceptionEntry(callerStaffID, mondayDate, "23:00", "02:00", false),
		},
	}
	s := &Service{
		repo:          repo,
		staffResolver: &mockStaffResolver{staffByEmployee: callerStaffID},
	}

	_, warnings, err := s.CreateShiftAssignment(context.Background(), testBusiness,
		domain.ShiftAssignmentRecord{ShiftID: 1, EmployeeID: 5, Date: mondayDate})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected a warning for an overnight block, got %v", warnings)
	}
}

// Approved leave still blocks. The soft-warning decision applies to availability
// only, and this guards against it being loosened by accident.
func TestCreateShiftAssignment_LeaveStillBlocks(t *testing.T) {
	repo := &mockRepo{
		shift: &domain.ShiftRecord{ID: 1, StartTime: "09:00:00", EndTime: "17:00:00"},
		forDate: []domain.AvailabilityRecord{
			recurringEntry(callerStaffID, 0, "08:00", "18:00", true),
		},
	}
	s := &Service{
		repo:                 repo,
		leaveConflictChecker: &mockLeaveChecker{hasLeave: true},
		staffResolver:        &mockStaffResolver{staffByEmployee: callerStaffID},
	}

	_, _, err := s.CreateShiftAssignment(context.Background(), testBusiness,
		domain.ShiftAssignmentRecord{ShiftID: 1, EmployeeID: 5, Date: mondayDate})
	if !errors.Is(err, domain.ErrRosterConflict) {
		t.Fatalf("approved leave must still block the assignment, got: %v", err)
	}
}
