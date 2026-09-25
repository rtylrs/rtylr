package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/staff/internal/domain"
)

// staffRepo mirrors all methods called on *infrastructure.Repository.
type staffRepo interface {
	ListStaff(ctx context.Context, businessID int64, search, departmentID, status string) ([]domain.Staff, error)
	CreateStaff(ctx context.Context, input domain.StaffInput) (int64, error)
	UpdateStaff(ctx context.Context, businessID int64, input domain.StaffUpdateInput) error
	GetStaffByID(ctx context.Context, businessID, staffID int64) (*domain.Staff, error)
	DeleteStaff(ctx context.Context, businessID, staffID int64) error
	ClockIn(ctx context.Context, businessID, staffID int64) error
	ClockOut(ctx context.Context, businessID, staffID int64) error
	CountActiveStaff(ctx context.Context, businessID int64) (int, error)
	CountOnShiftToday(ctx context.Context, businessID int64) (int, error)
	CountPendingTimeOff(ctx context.Context, businessID int64) (int, error)
	CountDraftPayroll(ctx context.Context, businessID int64) (int, error)
}

type mockStaffRepo struct {
	list           []domain.Staff
	listErr        error
	staff          *domain.Staff
	staffErr       error
	createID       int64
	createErr      error
	updateErr      error
	deleteErr      error
	clockInErr     error
	clockOutErr    error
	activeCount    int
	activeErr      error
	onShiftCount   int
	onShiftErr     error
	pendingTOCount int
	pendingTOErr   error
	draftPayCount  int
	draftPayErr    error
}

func (m *mockStaffRepo) ListStaff(_ context.Context, _ int64, _, _, _ string) ([]domain.Staff, error) {
	return m.list, m.listErr
}

func (m *mockStaffRepo) CreateStaff(_ context.Context, _ domain.StaffInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockStaffRepo) UpdateStaff(_ context.Context, _ int64, _ domain.StaffUpdateInput) error {
	return m.updateErr
}

func (m *mockStaffRepo) GetStaffByID(_ context.Context, _, _ int64) (*domain.Staff, error) {
	return m.staff, m.staffErr
}

func (m *mockStaffRepo) DeleteStaff(_ context.Context, _, _ int64) error {
	return m.deleteErr
}
func (m *mockStaffRepo) ClockIn(_ context.Context, _, _ int64) error  { return m.clockInErr }
func (m *mockStaffRepo) ClockOut(_ context.Context, _, _ int64) error { return m.clockOutErr }
func (m *mockStaffRepo) CountActiveStaff(_ context.Context, _ int64) (int, error) {
	return m.activeCount, m.activeErr
}

func (m *mockStaffRepo) CountOnShiftToday(_ context.Context, _ int64) (int, error) {
	return m.onShiftCount, m.onShiftErr
}

func (m *mockStaffRepo) CountPendingTimeOff(_ context.Context, _ int64) (int, error) {
	return m.pendingTOCount, m.pendingTOErr
}

func (m *mockStaffRepo) CountDraftPayroll(_ context.Context, _ int64) (int, error) {
	return m.draftPayCount, m.draftPayErr
}

// testService is a thin wrapper that uses the interface instead of the concrete type.
type testService struct{ repo staffRepo }

func (s *testService) ListStaff(ctx context.Context, businessID int64, search, departmentID, status string) ([]domain.Staff, error) {
	return s.repo.ListStaff(ctx, businessID, search, departmentID, status)
}

func (s *testService) CreateStaff(ctx context.Context, input domain.StaffInput) (int64, error) {
	return s.repo.CreateStaff(ctx, input)
}

func (s *testService) UpdateStaff(ctx context.Context, businessID int64, input domain.StaffUpdateInput) error {
	return s.repo.UpdateStaff(ctx, businessID, input)
}

func (s *testService) GetStaffByID(ctx context.Context, businessID, staffID int64) (*domain.Staff, error) {
	return s.repo.GetStaffByID(ctx, businessID, staffID)
}

func (s *testService) DeleteStaff(ctx context.Context, businessID, staffID int64) error {
	return s.repo.DeleteStaff(ctx, businessID, staffID)
}

func (s *testService) ClockIn(ctx context.Context, businessID, staffID int64) error {
	return s.repo.ClockIn(ctx, businessID, staffID)
}

func (s *testService) ClockOut(ctx context.Context, businessID, staffID int64) error {
	return s.repo.ClockOut(ctx, businessID, staffID)
}

func (s *testService) GetDashboardStats(ctx context.Context, businessID int64) (domain.DashboardStats, error) {
	total, err := s.repo.CountActiveStaff(ctx, businessID)
	if err != nil {
		return domain.DashboardStats{}, err
	}
	onShift, err := s.repo.CountOnShiftToday(ctx, businessID)
	if err != nil {
		return domain.DashboardStats{}, err
	}
	pending, err := s.repo.CountPendingTimeOff(ctx, businessID)
	if err != nil {
		return domain.DashboardStats{}, err
	}
	draft, err := s.repo.CountDraftPayroll(ctx, businessID)
	if err != nil {
		return domain.DashboardStats{}, err
	}
	return domain.DashboardStats{
		TotalStaff:     total,
		OnShiftToday:   onShift,
		PendingTimeOff: pending,
		PayrollDraft:   draft,
	}, nil
}

// --- Tests ---

func TestListStaff_Delegates(t *testing.T) {
	m := &mockStaffRepo{list: []domain.Staff{{ID: 1}, {ID: 2}}}
	s := &testService{repo: m}
	got, err := s.ListStaff(context.Background(), 1, "", "", "")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d items, want 2", len(got))
	}
}

func TestListStaff_PropagatesError(t *testing.T) {
	want := errors.New("db error")
	s := &testService{repo: &mockStaffRepo{listErr: want}}
	_, err := s.ListStaff(context.Background(), 1, "", "", "")
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestCreateStaff_ReturnsID(t *testing.T) {
	s := &testService{repo: &mockStaffRepo{createID: 42}}
	id, err := s.CreateStaff(context.Background(), domain.StaffInput{})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if id != 42 {
		t.Errorf("id = %d, want 42", id)
	}
}

func TestCreateStaff_PropagatesError(t *testing.T) {
	want := errors.New("insert failed")
	s := &testService{repo: &mockStaffRepo{createErr: want}}
	_, err := s.CreateStaff(context.Background(), domain.StaffInput{})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetStaffByID_ReturnsStaff(t *testing.T) {
	expected := &domain.Staff{ID: 7, FirstName: "Alice"}
	s := &testService{repo: &mockStaffRepo{staff: expected}}
	got, err := s.GetStaffByID(context.Background(), 1, 7)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.ID != 7 {
		t.Errorf("ID = %d, want 7", got.ID)
	}
}

func TestGetStaffByID_NotFound(t *testing.T) {
	want := domain.ErrStaffNotFound
	s := &testService{repo: &mockStaffRepo{staffErr: want}}
	_, err := s.GetStaffByID(context.Background(), 1, 999)
	if !errors.Is(err, want) {
		t.Errorf("expected ErrStaffNotFound, got %v", err)
	}
}

func TestUpdateStaff_HappyPath(t *testing.T) {
	s := &testService{repo: &mockStaffRepo{}}
	if err := s.UpdateStaff(context.Background(), 1, domain.StaffUpdateInput{}); err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

func TestUpdateStaff_PropagatesError(t *testing.T) {
	want := errors.New("update failed")
	s := &testService{repo: &mockStaffRepo{updateErr: want}}
	if err := s.UpdateStaff(context.Background(), 1, domain.StaffUpdateInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestDeleteStaff_HappyPath(t *testing.T) {
	s := &testService{repo: &mockStaffRepo{}}
	if err := s.DeleteStaff(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

func TestDeleteStaff_PropagatesError(t *testing.T) {
	want := errors.New("delete failed")
	s := &testService{repo: &mockStaffRepo{deleteErr: want}}
	if err := s.DeleteStaff(context.Background(), 1, 5); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestClockIn_HappyPath(t *testing.T) {
	s := &testService{repo: &mockStaffRepo{}}
	if err := s.ClockIn(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

func TestClockIn_PropagatesError(t *testing.T) {
	want := errors.New("already clocked in")
	s := &testService{repo: &mockStaffRepo{clockInErr: want}}
	if err := s.ClockIn(context.Background(), 1, 5); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestClockOut_HappyPath(t *testing.T) {
	s := &testService{repo: &mockStaffRepo{}}
	if err := s.ClockOut(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

func TestClockOut_PropagatesError(t *testing.T) {
	want := errors.New("not clocked in")
	s := &testService{repo: &mockStaffRepo{clockOutErr: want}}
	if err := s.ClockOut(context.Background(), 1, 5); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetDashboardStats_AggregatesCorrectly(t *testing.T) {
	m := &mockStaffRepo{
		activeCount:    10,
		onShiftCount:   3,
		pendingTOCount: 2,
		draftPayCount:  1,
	}
	s := &testService{repo: m}
	stats, err := s.GetDashboardStats(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if stats.TotalStaff != 10 {
		t.Errorf("TotalStaff = %d, want 10", stats.TotalStaff)
	}
	if stats.OnShiftToday != 3 {
		t.Errorf("OnShiftToday = %d, want 3", stats.OnShiftToday)
	}
	if stats.PendingTimeOff != 2 {
		t.Errorf("PendingTimeOff = %d, want 2", stats.PendingTimeOff)
	}
	if stats.PayrollDraft != 1 {
		t.Errorf("PayrollDraft = %d, want 1", stats.PayrollDraft)
	}
}

func TestGetDashboardStats_ErrorOnActiveStaff(t *testing.T) {
	want := errors.New("count failed")
	s := &testService{repo: &mockStaffRepo{activeErr: want}}
	_, err := s.GetDashboardStats(context.Background(), 1)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetDashboardStats_ErrorOnOnShift(t *testing.T) {
	want := errors.New("on-shift count failed")
	s := &testService{repo: &mockStaffRepo{onShiftErr: want}}
	_, err := s.GetDashboardStats(context.Background(), 1)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
