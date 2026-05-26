package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_shifts/internal/domain"
)

type mockRepo struct {
	shifts      []domain.ShiftRecord
	shift       *domain.ShiftRecord
	totalShifts int64
	assignments []domain.ShiftAssignmentRecord
	assignment  *domain.ShiftAssignmentRecord
	dateAssigns []domain.ShiftAssignmentRecord
	err         error
}

func (m *mockRepo) ListShifts(_ context.Context, _ int64, _ domain.ListShiftsParams) ([]domain.ShiftRecord, int64, error) {
	return m.shifts, m.totalShifts, m.err
}
func (m *mockRepo) GetShift(_ context.Context, _, _ int64) (*domain.ShiftRecord, error) {
	return m.shift, m.err
}
func (m *mockRepo) CreateShift(_ context.Context, _ int64, shift domain.ShiftRecord) (*domain.ShiftRecord, error) {
	return &shift, m.err
}
func (m *mockRepo) UpdateShift(_ context.Context, _ int64, _ domain.ShiftRecord) error { return m.err }
func (m *mockRepo) ListShiftsByLocation(_ context.Context, _, _ int64) ([]domain.ShiftRecord, error) {
	return m.shifts, m.err
}
func (m *mockRepo) ListRosterAssignments(_ context.Context, _ int64, _, _ string) ([]domain.ShiftAssignmentRecord, error) {
	return m.assignments, m.err
}
func (m *mockRepo) CreateShiftAssignment(_ context.Context, _ int64, a domain.ShiftAssignmentRecord) (*domain.ShiftAssignmentRecord, error) {
	return &a, m.err
}
func (m *mockRepo) PublishRosterAssignments(_ context.Context, _ int64, _, _ string) error {
	return m.err
}
func (m *mockRepo) ListEmployeeAssignmentsForDate(_ context.Context, _, _ int64, _ string) ([]domain.ShiftAssignmentRecord, error) {
	return m.dateAssigns, m.err
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

func TestCreateShift_DefaultsColor(t *testing.T) {
	s := &Service{repo: &mockRepo{shift: &domain.ShiftRecord{}}}
	got, err := s.CreateShift(context.Background(), 1, domain.ShiftRecord{Name: "Evening"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Color != "#4CAF50" {
		t.Errorf("expected default color, got %s", got.Color)
	}
}

func TestCreateShift_PreservesProvidedColor(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	got, err := s.CreateShift(context.Background(), 1, domain.ShiftRecord{Name: "Night", Color: "#FF0000"})
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
	_, err := s.CreateShift(context.Background(), 1, domain.ShiftRecord{Name: "Night"})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// ---- UpdateShift ----

func TestUpdateShift_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.UpdateShift(context.Background(), 1, domain.ShiftRecord{ID: 1, Name: "Updated"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateShift_PropagatesRepoError(t *testing.T) {
	want := errors.New("update error")
	s := &Service{repo: &mockRepo{err: want}}
	err := s.UpdateShift(context.Background(), 1, domain.ShiftRecord{ID: 1})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// ---- CreateShiftAssignment ----

func TestCreateShiftAssignment_PropagatesGetShiftError(t *testing.T) {
	want := domain.ErrShiftNotFound
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.CreateShiftAssignment(context.Background(), 1, domain.ShiftAssignmentRecord{ShiftID: 99})
	if !errors.Is(err, want) {
		t.Errorf("expected ErrShiftNotFound, got %v", err)
	}
}

func TestCreateShiftAssignment_LeaveConflictReturnsError(t *testing.T) {
	s := &Service{
		repo:                 &mockRepo{shift: &domain.ShiftRecord{ID: 1}},
		leaveConflictChecker: &mockLeaveChecker{hasLeave: true},
	}
	_, err := s.CreateShiftAssignment(context.Background(), 1, domain.ShiftAssignmentRecord{ShiftID: 1, EmployeeID: 5, Date: "2024-06-01"})
	if !errors.Is(err, domain.ErrRosterConflict) {
		t.Errorf("expected ErrRosterConflict, got %v", err)
	}
}

func TestCreateShiftAssignment_HappyPathNoConflicts(t *testing.T) {
	s := &Service{
		repo:                 &mockRepo{shift: &domain.ShiftRecord{ID: 1}},
		leaveConflictChecker: &mockLeaveChecker{hasLeave: false},
	}
	got, err := s.CreateShiftAssignment(context.Background(), 1, domain.ShiftAssignmentRecord{ShiftID: 1, EmployeeID: 5, Date: "2024-06-01"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ShiftID != 1 {
		t.Errorf("expected ShiftID=1, got %d", got.ShiftID)
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
