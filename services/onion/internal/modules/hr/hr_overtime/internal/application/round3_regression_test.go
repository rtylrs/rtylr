package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_overtime/internal/domain"
)

func TestCreateOvertimeRequest_RequiresManageEmployeesForAnotherSubject(t *testing.T) {
	repo := &mockRepo{}
	permissions := &mockPermissionProvider{
		employeeID: 101,
		businessID: 77,
		employeeBusinessID: map[int64]int64{
			101: 77,
			202: 77,
		},
		permissions: map[string]bool{
			domain.PermissionManageEmployees: false,
		},
	}
	service := &Service{repo: repo, permissions: permissions}

	_, err := service.CreateOvertimeRequest(context.Background(), 77, 9001, domain.OvertimeRequestRecord{EmployeeID: 202})
	if !errors.Is(err, domain.ErrOvertimeAccessDenied) {
		t.Fatalf("cross-employee create error = %v, want ErrOvertimeAccessDenied", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("repository create calls = %d, want 0", repo.createCalls)
	}

	permissions.permissions[domain.PermissionManageEmployees] = true
	if _, err := service.CreateOvertimeRequest(context.Background(), 77, 9001, domain.OvertimeRequestRecord{EmployeeID: 202}); err != nil {
		t.Fatalf("manage-employees create: %v", err)
	}
	if repo.createCalls != 1 {
		t.Fatalf("repository create calls = %d, want 1", repo.createCalls)
	}
}

func TestCreateOvertimeRequest_RejectsForeignEmployeeEvenWithPermission(t *testing.T) {
	repo := &mockRepo{}
	permissions := &mockPermissionProvider{
		employeeID: 101,
		businessID: 77,
		employeeBusinessID: map[int64]int64{
			101: 77,
			202: 88,
		},
		permissions: map[string]bool{
			domain.PermissionManageEmployees: true,
		},
	}
	service := &Service{repo: repo, permissions: permissions}

	_, err := service.CreateOvertimeRequest(context.Background(), 77, 9001, domain.OvertimeRequestRecord{EmployeeID: 202})
	if !errors.Is(err, domain.ErrOvertimeAccessDenied) {
		t.Fatalf("foreign employee error = %v, want ErrOvertimeAccessDenied", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("repository create calls = %d, want 0", repo.createCalls)
	}
}

func TestUpsertOvertimeConfig_RequiresManageHRConfig(t *testing.T) {
	repo := &mockRepo{}
	permissions := &mockPermissionProvider{
		employeeID: 101,
		businessID: 77,
		permissions: map[string]bool{
			domain.PermissionManageHRConfig: false,
		},
	}
	service := &Service{repo: repo, permissions: permissions}

	err := service.UpsertOvertimeConfig(context.Background(), 77, 9001, domain.OvertimeConfigRecord{
		AutoDetectEnabled: true,
	})
	if !errors.Is(err, domain.ErrOvertimeAccessDenied) {
		t.Fatalf("config mutation error = %v, want ErrOvertimeAccessDenied", err)
	}

	permissions.permissions[domain.PermissionManageHRConfig] = true
	if err := service.UpsertOvertimeConfig(context.Background(), 77, 9001, domain.OvertimeConfigRecord{}); err != nil {
		t.Fatalf("authorized config mutation: %v", err)
	}
}

func TestApproveOvertime_AdvancesIntermediateLevel(t *testing.T) {
	repo := &mockRepo{request: &domain.OvertimeRequestRecord{ID: 1, EmployeeID: 5, CurrentApprovalLevel: 0}}
	service := &Service{
		repo: repo,
		approvalProvider: &mockApprovalProvider{chain: &domain.ApprovalChainInfo{
			TotalLevels: 2,
			Levels: []domain.ApprovalLevelInfo{
				{LevelOrder: 1, ApproverEmployeeID: 2},
				{LevelOrder: 2, ApproverEmployeeID: 3},
			},
		}},
		permissions: &mockPermissionProvider{employeeID: 2, businessID: 77, allowed: true},
	}

	if err := service.ApproveOvertime(context.Background(), 77, 1, 9001, 1.5); err != nil {
		t.Fatalf("intermediate approval: %v", err)
	}
	if repo.advanceCalls != 1 {
		t.Fatalf("approval-level advances = %d, want 1", repo.advanceCalls)
	}
}

func TestRejectOvertime_RequiresAndPersistsReason(t *testing.T) {
	repo := &mockRepo{request: &domain.OvertimeRequestRecord{ID: 1, EmployeeID: 5}}
	service := &Service{
		repo: repo,
		approvalProvider: &mockApprovalProvider{chain: &domain.ApprovalChainInfo{
			TotalLevels: 1,
			Levels:      []domain.ApprovalLevelInfo{{LevelOrder: 1, ApproverEmployeeID: 2}},
		}},
		permissions: &mockPermissionProvider{employeeID: 2, businessID: 77, allowed: true},
	}

	err := service.RejectOvertime(context.Background(), 77, 1, 9001, "   ")
	if !errors.Is(err, domain.ErrOvertimeRejectionReason) {
		t.Fatalf("blank rejection reason error = %v, want ErrOvertimeRejectionReason", err)
	}
	if len(repo.rejectionReasons) != 0 {
		t.Fatalf("repository rejection calls = %d, want 0", len(repo.rejectionReasons))
	}

	if err := service.RejectOvertime(context.Background(), 77, 1, 9001, "  insufficient coverage  "); err != nil {
		t.Fatalf("authorized rejection: %v", err)
	}
	if len(repo.rejectionReasons) != 1 || repo.rejectionReasons[0] != "insufficient coverage" {
		t.Fatalf("persisted rejection reasons = %#v, want [\"insufficient coverage\"]", repo.rejectionReasons)
	}
}

func TestListOvertimeRequests_CacheKeySeparatesEmployeeFilters(t *testing.T) {
	repo := &mockRepo{}
	service := &Service{repo: repo}
	employeeOne := int64(1)
	employeeTwo := int64(2)

	if _, _, err := service.ListOvertimeRequests(context.Background(), 7802, domain.ListOvertimeRequestsParams{EmployeeID: &employeeOne, Limit: 10}); err != nil {
		t.Fatalf("first filtered list: %v", err)
	}
	if _, _, err := service.ListOvertimeRequests(context.Background(), 7802, domain.ListOvertimeRequestsParams{EmployeeID: &employeeTwo, Limit: 10}); err != nil {
		t.Fatalf("second filtered list: %v", err)
	}
	if repo.listCalls != 2 {
		t.Fatalf("repository list calls = %d, want 2 (filters collided in cache key)", repo.listCalls)
	}
}

func TestListOvertimeRequests_ForwardsEmployeeAndStatusFilters(t *testing.T) {
	repo := &mockRepo{requests: []domain.OvertimeRequestRecord{{ID: 1, EmployeeID: 42, Status: "pending"}}, totalReqs: 1}
	service := &Service{repo: repo}
	employeeID := int64(42)
	status := " PENDING "
	if _, _, err := service.ListOvertimeRequests(context.Background(), 7801, domain.ListOvertimeRequestsParams{
		EmployeeID: &employeeID,
		Status:     &status,
		Limit:      10,
	}); err != nil {
		t.Fatalf("filtered list: %v", err)
	}
	if repo.listParams.EmployeeID == nil || *repo.listParams.EmployeeID != 42 {
		t.Fatalf("forwarded employee filter = %#v, want 42", repo.listParams.EmployeeID)
	}
	if repo.listParams.Status == nil || *repo.listParams.Status != "pending" {
		t.Fatalf("forwarded status filter = %#v, want pending", repo.listParams.Status)
	}
}
