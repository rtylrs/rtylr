package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_overtime/internal/domain"
)

type mockRepo struct {
	requests         []domain.OvertimeRequestRecord
	request          *domain.OvertimeRequestRecord
	totalReqs        int64
	config           *domain.OvertimeConfigRecord
	ownerIDs         []int64
	flagged          []domain.OvertimeRequestRecord
	err              error
	configCalls      []int64
	listParams       domain.ListOvertimeRequestsParams
	listCalls        int
	createCalls      int
	advanceCalls     int
	rejectionReasons []string
}

func (m *mockRepo) ListOvertimeRequests(_ context.Context, _ int64, params domain.ListOvertimeRequestsParams) ([]domain.OvertimeRequestRecord, int64, error) {
	m.listCalls++
	m.listParams = params
	return m.requests, m.totalReqs, m.err
}

func (m *mockRepo) GetOvertimeRequest(_ context.Context, _, _ int64) (*domain.OvertimeRequestRecord, error) {
	return m.request, m.err
}

func (m *mockRepo) CreateOvertimeRequest(_ context.Context, _ int64, req domain.OvertimeRequestRecord) (*domain.OvertimeRequestRecord, error) {
	m.createCalls++
	return &req, m.err
}

func (m *mockRepo) ApproveOvertimeRequest(_ context.Context, _, _, _ int64, _ float64) error {
	return m.err
}

func (m *mockRepo) AdvanceOvertimeApprovalLevel(_ context.Context, _, _ int64, _ int32) error {
	m.advanceCalls++
	return m.err
}

func (m *mockRepo) RejectOvertimeRequest(_ context.Context, _, _ int64, rejectionReason string) error {
	m.rejectionReasons = append(m.rejectionReasons, rejectionReason)
	return m.err
}
func (m *mockRepo) CreateFlaggedOvertimeRequest(_ context.Context, _, _ int64, _ string, _ int32) error {
	return m.err
}

func (m *mockRepo) ListFlaggedForDate(_ context.Context, _ int64, _ string) ([]domain.OvertimeRequestRecord, error) {
	return m.flagged, m.err
}

func (m *mockRepo) ListOvertimeCandidates(_ context.Context, _ int64, _ string, _ int32) ([]domain.OvertimeDetectionCandidate, error) {
	return nil, m.err
}

func (m *mockRepo) GetOvertimeConfig(_ context.Context, businessID int64) (*domain.OvertimeConfigRecord, error) {
	m.configCalls = append(m.configCalls, businessID)
	return m.config, m.err
}

func (m *mockRepo) UpsertOvertimeConfig(_ context.Context, _ int64, _ domain.OvertimeConfigRecord) error {
	return m.err
}

func (m *mockRepo) ListAllConfiguredOwnerIDs(_ context.Context, limit, offset int32) ([]int64, error) {
	if m.err != nil {
		return nil, m.err
	}
	start := int(offset)
	if start >= len(m.ownerIDs) {
		return nil, nil
	}
	end := start + int(limit)
	if end > len(m.ownerIDs) {
		end = len(m.ownerIDs)
	}
	return m.ownerIDs[start:end], nil
}

type mockApprovalProvider struct {
	chain *domain.ApprovalChainInfo
	err   error
}

func (m *mockApprovalProvider) GetApprovalChain(_ context.Context, _ int64, _ string) (*domain.ApprovalChainInfo, error) {
	return m.chain, m.err
}

type mockPermissionProvider struct {
	employeeID         int64
	businessID         int64
	employeeBusinessID map[int64]int64
	allowed            bool
	permissions        map[string]bool
}

func (m *mockPermissionProvider) GetEmployeeIDByUserID(context.Context, int64, int64) (int64, error) {
	return m.employeeID, nil
}

func (m *mockPermissionProvider) GetEmployeeBusinessID(_ context.Context, _, employeeID int64) (int64, error) {
	if m.employeeBusinessID != nil {
		return m.employeeBusinessID[employeeID], nil
	}
	if m.businessID == 0 {
		return 1, nil
	}
	return m.businessID, nil
}

func (m *mockPermissionProvider) HasPermission(_ context.Context, _, _ int64, permissionKey string) (bool, error) {
	if m.permissions != nil {
		return m.permissions[permissionKey], nil
	}
	return m.allowed, nil
}

func (m *mockPermissionProvider) HasRole(context.Context, int64, int64, int64) (bool, error) {
	return m.allowed, nil
}

// ---- ListOvertimeRequests ----

func TestListOvertimeRequests_ReturnsDelegatedItems(t *testing.T) {
	items := []domain.OvertimeRequestRecord{{ID: 1, Status: "pending"}}
	s := &Service{repo: &mockRepo{requests: items, totalReqs: 1}}
	got, total, err := s.ListOvertimeRequests(context.Background(), 1, domain.ListOvertimeRequestsParams{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || total != 1 {
		t.Errorf("expected 1 item, got %d (total=%d)", len(got), total)
	}
}

func TestListOvertimeRequests_ReturnsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	got, _, err := s.ListOvertimeRequests(context.Background(), 1, domain.ListOvertimeRequestsParams{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %+v", got)
	}
}

func TestListOvertimeRequests_PropagatesRepoError(t *testing.T) {
	want := errors.New("db error")
	s := &Service{repo: &mockRepo{err: want}}
	_, _, err := s.ListOvertimeRequests(context.Background(), 1, domain.ListOvertimeRequestsParams{Limit: 10})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// ---- CreateOvertimeRequest ----

func TestCreateOvertimeRequest_SetsPendingStatus(t *testing.T) {
	s := &Service{repo: &mockRepo{}, permissions: &mockPermissionProvider{employeeID: 5, businessID: 1}}
	got, err := s.CreateOvertimeRequest(context.Background(), 1, 77, domain.OvertimeRequestRecord{EmployeeID: 5, Status: "approved"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != "pending" {
		t.Errorf("expected status=pending, got %s", got.Status)
	}
}

func TestCreateOvertimeRequest_SetsApprovalLevelZero(t *testing.T) {
	s := &Service{repo: &mockRepo{}, permissions: &mockPermissionProvider{employeeID: 5, businessID: 1}}
	got, err := s.CreateOvertimeRequest(context.Background(), 1, 77, domain.OvertimeRequestRecord{EmployeeID: 5, CurrentApprovalLevel: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.CurrentApprovalLevel != 0 {
		t.Errorf("expected CurrentApprovalLevel=0, got %d", got.CurrentApprovalLevel)
	}
}

func TestCreateOvertimeRequest_PropagatesRepoError(t *testing.T) {
	want := errors.New("create error")
	s := &Service{repo: &mockRepo{err: want}, permissions: &mockPermissionProvider{employeeID: 5, businessID: 1}}
	_, err := s.CreateOvertimeRequest(context.Background(), 1, 77, domain.OvertimeRequestRecord{EmployeeID: 5})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// ---- ApproveOvertime ----

func TestApproveOvertime_PropagatesGetRequestError(t *testing.T) {
	want := domain.ErrOvertimeRequestNotFound
	s := &Service{repo: &mockRepo{err: want}, permissions: &mockPermissionProvider{employeeID: 2, allowed: true}}
	err := s.ApproveOvertime(context.Background(), 1, 99, 1, 1.5)
	if !errors.Is(err, want) {
		t.Errorf("expected ErrOvertimeRequestNotFound, got %v", err)
	}
}

func TestApproveOvertime_ApprovalChainNotConfiguredReturnsError(t *testing.T) {
	req := &domain.OvertimeRequestRecord{ID: 1, CurrentApprovalLevel: 0}
	s := &Service{
		repo:             &mockRepo{request: req},
		approvalProvider: &mockApprovalProvider{chain: nil},
		permissions:      &mockPermissionProvider{employeeID: 2, allowed: true},
	}
	err := s.ApproveOvertime(context.Background(), 1, 1, 1, 1.5)
	if !errors.Is(err, domain.ErrApprovalChainNotConfigured) {
		t.Errorf("expected ErrApprovalChainNotConfigured, got %v", err)
	}
}

func TestApproveOvertime_FinalLevelApproves(t *testing.T) {
	req := &domain.OvertimeRequestRecord{ID: 1, EmployeeID: 5, CurrentApprovalLevel: 0}
	chain := &domain.ApprovalChainInfo{TotalLevels: 1, Levels: []domain.ApprovalLevelInfo{{LevelOrder: 1, ApproverEmployeeID: 2}}}
	s := &Service{
		repo:             &mockRepo{request: req, config: &domain.OvertimeConfigRecord{DefaultMultiplier: 2.0}},
		approvalProvider: &mockApprovalProvider{chain: chain},
		permissions:      &mockPermissionProvider{employeeID: 2, allowed: true},
	}
	err := s.ApproveOvertime(context.Background(), 1, 1, 1, 1.5)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestApproveOvertime_RejectsSelfApproval(t *testing.T) {
	s := &Service{
		repo:        &mockRepo{request: &domain.OvertimeRequestRecord{ID: 1, EmployeeID: 2}},
		permissions: &mockPermissionProvider{employeeID: 2, allowed: true},
		approvalProvider: &mockApprovalProvider{chain: &domain.ApprovalChainInfo{
			TotalLevels: 1,
			Levels:      []domain.ApprovalLevelInfo{{LevelOrder: 1, ApproverEmployeeID: 2}},
		}},
	}

	err := s.ApproveOvertime(context.Background(), 1, 1, 77, 1.5)
	if !errors.Is(err, domain.ErrOvertimeSelfApproval) {
		t.Fatalf("expected self-approval denial, got %v", err)
	}
}

func TestApproveOvertime_RequiresConfiguredMultiplier(t *testing.T) {
	s := &Service{
		repo:        &mockRepo{request: &domain.OvertimeRequestRecord{ID: 1, EmployeeID: 5}},
		permissions: &mockPermissionProvider{employeeID: 2, allowed: true},
		approvalProvider: &mockApprovalProvider{chain: &domain.ApprovalChainInfo{
			TotalLevels: 1,
			Levels:      []domain.ApprovalLevelInfo{{LevelOrder: 1, ApproverEmployeeID: 2}},
		}},
	}

	err := s.ApproveOvertime(context.Background(), 1, 1, 77, 0)
	if !errors.Is(err, domain.ErrOvertimeConfigNotFound) {
		t.Fatalf("expected missing-config error, got %v", err)
	}
}

// ---- RejectOvertime ----

func TestRejectOvertime_HappyPath(t *testing.T) {
	req := &domain.OvertimeRequestRecord{ID: 1, EmployeeID: 5}
	s := &Service{repo: &mockRepo{request: req}, approvalProvider: &mockApprovalProvider{chain: &domain.ApprovalChainInfo{TotalLevels: 1, Levels: []domain.ApprovalLevelInfo{{LevelOrder: 1, ApproverEmployeeID: 2}}}}, permissions: &mockPermissionProvider{employeeID: 2, allowed: true}}
	err := s.RejectOvertime(context.Background(), 1, 1, 77, "insufficient coverage")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRejectOvertime_PropagatesGetRequestError(t *testing.T) {
	want := domain.ErrOvertimeRequestNotFound
	s := &Service{repo: &mockRepo{err: want}, permissions: &mockPermissionProvider{employeeID: 2, allowed: true}}
	err := s.RejectOvertime(context.Background(), 1, 99, 77, "insufficient coverage")
	if !errors.Is(err, want) {
		t.Errorf("expected ErrOvertimeRequestNotFound, got %v", err)
	}
}

// ---- GetOvertimeConfig ----

func TestGetOvertimeConfig_PropagatesNotFoundError(t *testing.T) {
	s := &Service{repo: &mockRepo{err: domain.ErrOvertimeConfigNotFound}}
	_, err := s.GetOvertimeConfig(context.Background(), 1)
	if !errors.Is(err, domain.ErrOvertimeConfigNotFound) {
		t.Errorf("expected ErrOvertimeConfigNotFound, got %v", err)
	}
}

func TestGetOvertimeConfig_PropagatesUnexpectedRepoError(t *testing.T) {
	want := errors.New("connection reset")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.GetOvertimeConfig(context.Background(), 1)
	if !errors.Is(err, want) {
		t.Errorf("expected the read failure to surface, got %v", err)
	}
	if errors.Is(err, domain.ErrOvertimeConfigNotFound) {
		t.Errorf("transient read failure was mapped to ErrOvertimeConfigNotFound")
	}
}

func TestGetOvertimeConfig_ReturnsDelegatedConfig(t *testing.T) {
	cfg := &domain.OvertimeConfigRecord{DefaultMultiplier: 1.5, AutoDetectEnabled: true}
	s := &Service{repo: &mockRepo{config: cfg}}
	got, err := s.GetOvertimeConfig(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.DefaultMultiplier != 1.5 {
		t.Errorf("expected 1.5, got %f", got.DefaultMultiplier)
	}
}

// ---- UpsertOvertimeConfig ----

func TestUpsertOvertimeConfig_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}, permissions: &mockPermissionProvider{employeeID: 5, businessID: 1, permissions: map[string]bool{domain.PermissionManageHRConfig: true}}}
	err := s.UpsertOvertimeConfig(context.Background(), 1, 77, domain.OvertimeConfigRecord{DefaultMultiplier: 2.0})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpsertOvertimeConfig_PropagatesRepoError(t *testing.T) {
	want := errors.New("upsert error")
	s := &Service{repo: &mockRepo{err: want}, permissions: &mockPermissionProvider{employeeID: 5, businessID: 1, permissions: map[string]bool{domain.PermissionManageHRConfig: true}}}
	err := s.UpsertOvertimeConfig(context.Background(), 1, 77, domain.OvertimeConfigRecord{})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// ---- DetectAndFlagOvertime ----

func TestDetectAndFlagOvertime_MissingConfigSkipsSilently(t *testing.T) {
	s := &Service{repo: &mockRepo{err: domain.ErrOvertimeConfigNotFound}}
	if err := s.DetectAndFlagOvertime(context.Background(), 1, "2026-03-20"); err != nil {
		t.Fatalf("missing config should skip, got %v", err)
	}
}

func TestDetectAndFlagOvertime_PropagatesConfigReadError(t *testing.T) {
	want := errors.New("connection reset")
	s := &Service{repo: &mockRepo{err: want}}
	err := s.DetectAndFlagOvertime(context.Background(), 1, "2026-03-20")
	if !errors.Is(err, want) {
		t.Fatalf("transient config read error was swallowed, got %v", err)
	}
}

// ---- RunDailyDetection ----

func TestRunDailyDetection_PaginatesAcrossAllConfiguredOwners(t *testing.T) {
	ownerIDs := make([]int64, 250)
	for i := range ownerIDs {
		ownerIDs[i] = int64(i + 1)
	}
	repo := &mockRepo{ownerIDs: ownerIDs, config: &domain.OvertimeConfigRecord{AutoDetectEnabled: false}}
	s := &Service{repo: repo}

	if err := s.RunDailyDetection(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.configCalls) != len(ownerIDs) {
		t.Fatalf("expected every configured owner to be visited across pages, got %d of %d",
			len(repo.configCalls), len(ownerIDs))
	}
}

func TestRunDailyDetection_PropagatesListOwnerIDsError(t *testing.T) {
	want := errors.New("list owner ids error")
	s := &Service{repo: &mockRepo{err: want}}
	err := s.RunDailyDetection(context.Background())
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
