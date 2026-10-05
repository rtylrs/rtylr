package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/domain"
)

type mockRepo struct {
	shifts      []domain.ShiftRecord
	shift       *domain.ShiftRecord
	totalShifts int64
	assignments []domain.ShiftAssignmentRecord
	dateAssigns []domain.ShiftAssignmentRecord
	err         error

	lastParams  domain.ListShiftsParams
	updated     domain.ShiftRecord
	deletedID   int64
	deleteCalls int

	createAssignCalls int

	availability      []domain.AvailabilityRecord
	availabilityTotal int64
	availabilityOne   *domain.AvailabilityRecord
	forDate           []domain.AvailabilityRecord
	availabilityErr   error

	lastAvailParams  domain.ListAvailabilityParams
	upserted         domain.AvailabilityRecord
	deletedAvailID   int64
	availDeleteCalls int
	forDateStaffID   int64
	forDateDayOfWeek int32
	forDateOnDate    string
	forDateCalls     int
}

func (m *mockRepo) ListShifts(_ context.Context, _ int64, params domain.ListShiftsParams) ([]domain.ShiftRecord, int64, error) {
	m.lastParams = params
	return m.shifts, m.totalShifts, m.err
}

func (m *mockRepo) GetShift(_ context.Context, _, _ int64) (*domain.ShiftRecord, error) {
	return m.shift, m.err
}

func (m *mockRepo) CreateShift(_ context.Context, _ int64, shift domain.ShiftRecord) (*domain.ShiftRecord, error) {
	return &shift, m.err
}

func (m *mockRepo) UpdateShift(_ context.Context, _ int64, shift domain.ShiftRecord) error {
	m.updated = shift
	return m.err
}

func (m *mockRepo) DeleteShift(_ context.Context, _, shiftID int64) error {
	m.deletedID = shiftID
	m.deleteCalls++
	return m.err
}

func (m *mockRepo) ListShiftsByLocation(_ context.Context, _, _ int64) ([]domain.ShiftRecord, error) {
	return m.shifts, m.err
}

func (m *mockRepo) ListRosterAssignments(_ context.Context, _ int64, _, _ string) ([]domain.ShiftAssignmentRecord, error) {
	return m.assignments, m.err
}

func (m *mockRepo) CreateShiftAssignment(_ context.Context, _ int64, a domain.ShiftAssignmentRecord) (*domain.ShiftAssignmentRecord, error) {
	m.createAssignCalls++
	return &a, m.err
}

func (m *mockRepo) PublishRosterAssignments(_ context.Context, _ int64, _, _ string) error {
	return m.err
}

func (m *mockRepo) ListEmployeeAssignmentsForDate(_ context.Context, _, _ int64, _ string) ([]domain.ShiftAssignmentRecord, error) {
	return m.dateAssigns, m.err
}

func (m *mockRepo) ListAvailability(_ context.Context, _ int64, params domain.ListAvailabilityParams) ([]domain.AvailabilityRecord, int64, error) {
	m.lastAvailParams = params
	return m.availability, m.availabilityTotal, m.availabilityErr
}

func (m *mockRepo) GetAvailability(_ context.Context, _, _ int64) (*domain.AvailabilityRecord, error) {
	if m.availabilityOne == nil && m.availabilityErr == nil {
		return nil, domain.ErrAvailabilityNotFound
	}
	return m.availabilityOne, m.availabilityErr
}

func (m *mockRepo) UpsertAvailability(_ context.Context, _ int64, entry domain.AvailabilityRecord) (*domain.AvailabilityRecord, error) {
	m.upserted = entry
	if m.availabilityErr != nil {
		return nil, m.availabilityErr
	}
	saved := entry
	return &saved, nil
}

func (m *mockRepo) DeleteAvailability(_ context.Context, _, id int64) error {
	m.deletedAvailID = id
	m.availDeleteCalls++
	return m.availabilityErr
}

func (m *mockRepo) ListAvailabilityForStaffDate(_ context.Context, _, staffID int64, date string, dayOfWeek int32) ([]domain.AvailabilityRecord, error) {
	m.forDateCalls++
	m.forDateStaffID = staffID
	m.forDateOnDate = date
	m.forDateDayOfWeek = dayOfWeek
	return m.forDate, m.availabilityErr
}

// mockStaffResolver stands in for the hr_staff lookups. staffByUser and
// staffByEmployee are separate because those are genuinely different identities.
type mockStaffResolver struct {
	staffByUser     int64
	staffByEmployee int64
	userErr         error
	employeeErr     error
}

func (m *mockStaffResolver) StaffIDByUserID(_ context.Context, _, _ int64) (int64, error) {
	return m.staffByUser, m.userErr
}

func (m *mockStaffResolver) StaffIDByEmployeeID(_ context.Context, _, _ int64) (int64, error) {
	return m.staffByEmployee, m.employeeErr
}

// mockPermissions mirrors the provider hr_payroll gates on.
type mockPermissions struct {
	employeeID  int64
	allowed     bool
	employeeErr error
	permErr     error
}

func (m *mockPermissions) GetEmployeeIDByUserID(_ context.Context, _, _ int64) (int64, error) {
	return m.employeeID, m.employeeErr
}

func (m *mockPermissions) HasPermission(_ context.Context, _, _ int64, _ string) (bool, error) {
	return m.allowed, m.permErr
}

type mockLeaveChecker struct {
	hasLeave bool
	err      error
}

func (m *mockLeaveChecker) HasApprovedLeave(_ context.Context, _, _ int64, _ string) (bool, error) {
	return m.hasLeave, m.err
}

// ---- ListShifts ----

func TestListShifts_ReturnsDelegatedItems(t *testing.T) {
	items := []domain.ShiftRecord{{ID: 1, Name: "Morning"}}
	s := &Service{repo: &mockRepo{shifts: items, totalShifts: 1}}
	got, total, err := s.ListShifts(context.Background(), 1, domain.ListShiftsParams{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || total != 1 {
		t.Errorf("expected 1 item, got %d (total=%d)", len(got), total)
	}
}

func TestListShifts_ReturnsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	got, _, err := s.ListShifts(context.Background(), 1, domain.ListShiftsParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %+v", got)
	}
}

func TestListShifts_PropagatesRepoError(t *testing.T) {
	want := errors.New("db error")
	s := &Service{repo: &mockRepo{err: want}}
	_, _, err := s.ListShifts(context.Background(), 1, domain.ListShiftsParams{Limit: 10})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// ---- CreateShift ----

// templateShift is a valid template row for tests: a location and name, no staff.
func templateShift() domain.ShiftRecord {
	return domain.ShiftRecord{
		LocationID: 7,
		Name:       "Evening",
		StartTime:  "18:00:00",
		EndTime:    "23:00:00",
	}
}

// attendanceShift is a valid attendance row for tests: a staff member and date.
func attendanceShift() domain.ShiftRecord {
	return domain.ShiftRecord{
		StaffID:   4101,
		Date:      "2026-06-01",
		StartTime: "09:00:00",
		EndTime:   "17:00:00",
	}
}

func TestCreateShift_DefaultsColor(t *testing.T) {
	s := &Service{repo: &mockRepo{shift: &domain.ShiftRecord{}}}
	got, err := s.CreateShift(context.Background(), 1, templateShift())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Color != domain.DefaultShiftColor {
		t.Errorf("expected default color, got %s", got.Color)
	}
}

func TestCreateShift_DefaultsStatus(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	got, err := s.CreateShift(context.Background(), 1, attendanceShift())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != domain.DefaultShiftStatus {
		t.Errorf("expected default status, got %q", got.Status)
	}
}

func TestCreateShift_PreservesProvidedColor(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	shift := templateShift()
	shift.Color = "#FF0000"
	got, err := s.CreateShift(context.Background(), 1, shift)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Color != "#FF0000" {
		t.Errorf("expected #FF0000, got %s", got.Color)
	}
}

func TestCreateShift_PropagatesRepoError(t *testing.T) {
	want := errors.New("create error")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.CreateShift(context.Background(), 1, templateShift())
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// A template row has no staff member, so it needs a name to be identifiable.
func TestCreateShift_RejectsTemplateWithoutName(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	shift := templateShift()
	shift.Name = ""
	if _, err := s.CreateShift(context.Background(), 1, shift); !errors.Is(err, domain.ErrEmptyShiftName) {
		t.Errorf("expected ErrEmptyShiftName, got %v", err)
	}
}

// An attendance row is a specific staff member on a specific day, so the date
// is required even though a name is not.
func TestCreateShift_RejectsAttendanceWithoutDate(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	shift := attendanceShift()
	shift.Date = ""
	if _, err := s.CreateShift(context.Background(), 1, shift); !errors.Is(err, domain.ErrEmptyShiftDate) {
		t.Errorf("expected ErrEmptyShiftDate, got %v", err)
	}
}

// An attendance row carries no name, so it must not be rejected for missing one.
func TestCreateShift_AcceptsAttendanceWithoutName(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if _, err := s.CreateShift(context.Background(), 1, attendanceShift()); err != nil {
		t.Errorf("expected attendance row without a name to be accepted, got %v", err)
	}
}

// ---- ListShifts filters ----

func TestListShifts_PassesFiltersThrough(t *testing.T) {
	repo := &mockRepo{}
	s := &Service{repo: repo}
	want := domain.ListShiftsParams{
		Limit: 10, Offset: 5, StaffID: 4101, Date: "2026-06-01",
		LocationID: 7, Kind: domain.ShiftKindAttendance,
	}
	if _, _, err := s.ListShifts(context.Background(), 1, want); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastParams != want {
		t.Errorf("filters not passed through: got %+v, want %+v", repo.lastParams, want)
	}
}

// ---- UpdateShift ----

func TestUpdateShift_HappyPath(t *testing.T) {
	stored := templateShift()
	stored.ID = 1
	s := &Service{repo: &mockRepo{shift: &stored}}
	got, err := s.UpdateShift(context.Background(), 1, domain.ShiftRecord{ID: 1, Name: "Updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected the updated shift to be returned")
	}
}

// A partial update must not blank the fields it leaves out. This is the failure
// mode where a caller sending only one field wipes the rest of the row.
func TestUpdateShift_PartialEditPreservesOtherFields(t *testing.T) {
	stored := attendanceShift()
	stored.ID = 42
	stored.Notes = "original note"
	stored.DepartmentID = 9
	stored.BreakDurationMinutes = 30

	repo := &mockRepo{shift: &stored}
	s := &Service{repo: repo}

	if _, err := s.UpdateShift(context.Background(), 1, domain.ShiftRecord{ID: 42, Status: "completed"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.updated.Status != "completed" {
		t.Errorf("Status = %q, want completed", repo.updated.Status)
	}
	if repo.updated.Notes != "original note" {
		t.Errorf("Notes = %q, want it preserved", repo.updated.Notes)
	}
	if repo.updated.DepartmentID != 9 {
		t.Errorf("DepartmentID = %d, want 9 preserved", repo.updated.DepartmentID)
	}
	if repo.updated.BreakDurationMinutes != 30 {
		t.Errorf("BreakDurationMinutes = %d, want 30 preserved", repo.updated.BreakDurationMinutes)
	}
	if repo.updated.StartTime != stored.StartTime {
		t.Errorf("StartTime = %q, want %q preserved", repo.updated.StartTime, stored.StartTime)
	}
}

// StaffID decides whether a row is a template or an attendance record, so an
// update must never flip one into the other.
func TestUpdateShift_DoesNotChangeShiftKind(t *testing.T) {
	stored := attendanceShift()
	stored.ID = 42
	repo := &mockRepo{shift: &stored}
	s := &Service{repo: repo}

	if _, err := s.UpdateShift(context.Background(), 1, domain.ShiftRecord{ID: 42, StaffID: 999, Name: "Sneaky"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.updated.StaffID != stored.StaffID {
		t.Errorf("StaffID = %d, want %d unchanged", repo.updated.StaffID, stored.StaffID)
	}
	if repo.updated.Kind() != domain.ShiftKindAttendance {
		t.Errorf("kind = %q, want it to stay attendance", repo.updated.Kind())
	}
}

func TestUpdateShift_MissingShiftReturnsNotFound(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if _, err := s.UpdateShift(context.Background(), 1, domain.ShiftRecord{ID: 7}); !errors.Is(err, domain.ErrShiftNotFound) {
		t.Errorf("expected ErrShiftNotFound, got %v", err)
	}
}

func TestUpdateShift_PropagatesRepoError(t *testing.T) {
	want := errors.New("update error")
	s := &Service{repo: &mockRepo{err: want}}
	if _, err := s.UpdateShift(context.Background(), 1, domain.ShiftRecord{ID: 1}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// ---- DeleteShift ----

func TestDeleteShift_DelegatesID(t *testing.T) {
	repo := &mockRepo{}
	s := &Service{repo: repo}
	if err := s.DeleteShift(context.Background(), 1, 55); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.deletedID != 55 {
		t.Errorf("deleted id = %d, want 55", repo.deletedID)
	}
	if repo.deleteCalls != 1 {
		t.Errorf("delete called %d times, want 1", repo.deleteCalls)
	}
}

func TestDeleteShift_PropagatesRepoError(t *testing.T) {
	want := errors.New("delete error")
	s := &Service{repo: &mockRepo{err: want}}
	if err := s.DeleteShift(context.Background(), 1, 55); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// ---- Kind discrimination ----

func TestShiftKind_StaffIDDiscriminates(t *testing.T) {
	if got := templateShift().Kind(); got != domain.ShiftKindTemplate {
		t.Errorf("template kind = %q, want template", got)
	}
	if got := attendanceShift().Kind(); got != domain.ShiftKindAttendance {
		t.Errorf("attendance kind = %q, want attendance", got)
	}
}

// ---- CreateShiftAssignment ----

func TestCreateShiftAssignment_PropagatesGetShiftError(t *testing.T) {
	want := domain.ErrShiftNotFound
	s := &Service{repo: &mockRepo{err: want}}
	_, _, err := s.CreateShiftAssignment(context.Background(), 1, domain.ShiftAssignmentRecord{ShiftID: 99})
	if !errors.Is(err, want) {
		t.Errorf("expected ErrShiftNotFound, got %v", err)
	}
}

func TestCreateShiftAssignment_LeaveConflictReturnsError(t *testing.T) {
	s := &Service{
		repo:                 &mockRepo{shift: &domain.ShiftRecord{ID: 1}},
		leaveConflictChecker: &mockLeaveChecker{hasLeave: true},
	}
	_, _, err := s.CreateShiftAssignment(context.Background(), 1, domain.ShiftAssignmentRecord{ShiftID: 1, EmployeeID: 5, Date: "2024-06-01"})
	if !errors.Is(err, domain.ErrRosterConflict) {
		t.Errorf("expected ErrRosterConflict, got %v", err)
	}
}

func TestCreateShiftAssignment_HappyPathNoConflicts(t *testing.T) {
	s := &Service{
		repo:                 &mockRepo{shift: &domain.ShiftRecord{ID: 1}},
		leaveConflictChecker: &mockLeaveChecker{hasLeave: false},
	}
	got, _, err := s.CreateShiftAssignment(context.Background(), 1, domain.ShiftAssignmentRecord{ShiftID: 1, EmployeeID: 5, Date: "2024-06-01"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ShiftID != 1 {
		t.Errorf("expected ShiftID=1, got %d", got.ShiftID)
	}
}

// A second assignment for the same employee and date collides with
// uk_shift_assignment, so it has to be reported as the documented roster
// conflict instead of falling through to a duplicate-key 500.
func TestCreateShiftAssignment_ExistingShiftReturnsConflict(t *testing.T) {
	repo := &mockRepo{
		shift:       &domain.ShiftRecord{ID: 1},
		dateAssigns: []domain.ShiftAssignmentRecord{{ID: 7, EmployeeID: 5, Date: "2024-06-01"}},
	}
	s := &Service{
		repo:                 repo,
		leaveConflictChecker: &mockLeaveChecker{hasLeave: false},
	}

	_, _, err := s.CreateShiftAssignment(context.Background(), 1, domain.ShiftAssignmentRecord{ShiftID: 1, EmployeeID: 5, Date: "2024-06-01"})
	if !errors.Is(err, domain.ErrRosterConflict) {
		t.Errorf("expected ErrRosterConflict, got %v", err)
	}
	if repo.createAssignCalls != 0 {
		t.Errorf("conflicting assignment was inserted %d time(s)", repo.createAssignCalls)
	}
}

// ---- GetRoster ----

func TestGetRoster_InvalidWeekStartReturnsError(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	_, _, err := s.GetRoster(context.Background(), 1, domain.RosterParams{WeekStart: "not-a-date"})
	if err == nil {
		t.Error("expected parse error, got nil")
	}
}

func TestGetRoster_ReturnsDelegatedAssignments(t *testing.T) {
	assigns := []domain.ShiftAssignmentRecord{{ID: 1, EmployeeID: 5, Date: "2024-06-03"}}
	s := &Service{repo: &mockRepo{assignments: assigns}}
	got, _, err := s.GetRoster(context.Background(), 1, domain.RosterParams{WeekStart: "2024-06-03"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 assignment, got %d", len(got))
	}
}

// ---- PublishRoster ----

func TestPublishRoster_InvalidWeekStartReturnsError(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.PublishRoster(context.Background(), 1, 1, "bad-date")
	if err == nil {
		t.Error("expected parse error, got nil")
	}
}

func TestPublishRoster_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.PublishRoster(context.Background(), 1, 1, "2024-06-03")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---- DetectConflicts ----

func TestDetectConflicts_NoConflictsWhenNoLeaveAndNoExistingShifts(t *testing.T) {
	s := &Service{
		repo:                 &mockRepo{},
		leaveConflictChecker: &mockLeaveChecker{hasLeave: false},
	}
	conflicts := s.DetectConflicts(context.Background(), 1, 5, "2024-06-01")
	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts, got %+v", conflicts)
	}
}

func TestDetectConflicts_LeaveConflictDetected(t *testing.T) {
	s := &Service{
		repo:                 &mockRepo{},
		leaveConflictChecker: &mockLeaveChecker{hasLeave: true},
	}
	conflicts := s.DetectConflicts(context.Background(), 1, 5, "2024-06-01")
	if len(conflicts) != 1 || conflicts[0].ConflictType != "leave" {
		t.Errorf("expected leave conflict, got %+v", conflicts)
	}
}
