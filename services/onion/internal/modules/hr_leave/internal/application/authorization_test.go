package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
)

type fakeLeavePermissionProvider struct {
	employeeID    int64
	employeeErr   error
	canApprove    bool
	permissionErr error
	roles         map[int64]bool
	roleErr       error
}

func (f fakeLeavePermissionProvider) GetEmployeeIDByUserID(context.Context, int64, int64) (int64, error) {
	return f.employeeID, f.employeeErr
}

func (f fakeLeavePermissionProvider) HasPermission(context.Context, int64, int64, string) (bool, error) {
	return f.canApprove, f.permissionErr
}

func (f fakeLeavePermissionProvider) HasRole(_ context.Context, _, _, roleID int64) (bool, error) {
	return f.roles[roleID], f.roleErr
}

func TestScopeLeaveRequests_ForcesCurrentEmployeeForNonApprover(t *testing.T) {
	svc := &Service{permissions: fakeLeavePermissionProvider{employeeID: 4101}}

	got, err := svc.scopeLeaveRequests(context.Background(), 1001, 77, 4102)
	if err != nil {
		t.Fatalf("scopeLeaveRequests: %v", err)
	}
	if got != 4101 {
		t.Fatalf("employee scope = %d, want current employee 4101", got)
	}
}

func TestScopeLeaveRequests_AllowsRequestedTeamScopeForApprover(t *testing.T) {
	svc := &Service{permissions: fakeLeavePermissionProvider{employeeID: 4101, canApprove: true}}

	got, err := svc.scopeLeaveRequests(context.Background(), 1001, 77, 4102)
	if err != nil {
		t.Fatalf("scopeLeaveRequests: %v", err)
	}
	if got != 4102 {
		t.Fatalf("employee scope = %d, want requested employee 4102", got)
	}

	got, err = svc.scopeLeaveRequests(context.Background(), 1001, 77, 0)
	if err != nil {
		t.Fatalf("scopeLeaveRequests without filter: %v", err)
	}
	if got != 0 {
		t.Fatalf("unfiltered approver scope = %d, want 0 for team feed", got)
	}
}

func TestScopeLeaveRequests_FailsClosedWithoutEmployee(t *testing.T) {
	svc := &Service{permissions: fakeLeavePermissionProvider{
		employeeErr: errors.New("employee lookup failed"),
	}}

	_, err := svc.scopeLeaveRequests(context.Background(), 1001, 77, 4102)
	if !errors.Is(err, domain.ErrLeaveAccessDenied) {
		t.Fatalf("error = %v, want ErrLeaveAccessDenied", err)
	}
}

func TestAuthorizeLeaveSubject_AllowsOwnEmployee(t *testing.T) {
	svc := &Service{permissions: fakeLeavePermissionProvider{employeeID: 4101}}

	if err := svc.authorizeLeaveSubject(context.Background(), 1001, 77, 4101); err != nil {
		t.Fatalf("authorizeLeaveSubject for own employee: %v", err)
	}
}

func TestAuthorizeLeaveSubject_DeniesOtherEmployeeWithoutApprovalPermission(t *testing.T) {
	svc := &Service{permissions: fakeLeavePermissionProvider{employeeID: 4101}}

	err := svc.authorizeLeaveSubject(context.Background(), 1001, 77, 4102)
	if !errors.Is(err, domain.ErrLeaveAccessDenied) {
		t.Fatalf("error = %v, want ErrLeaveAccessDenied", err)
	}
}

func TestAuthorizeLeaveSubject_AllowsOtherEmployeeForApprover(t *testing.T) {
	svc := &Service{permissions: fakeLeavePermissionProvider{employeeID: 4101, canApprove: true}}

	if err := svc.authorizeLeaveSubject(context.Background(), 1001, 77, 4102); err != nil {
		t.Fatalf("authorizeLeaveSubject for approver acting on someone else: %v", err)
	}
}

func TestAuthorizeLeaveSubject_FailsClosedWithoutEmployee(t *testing.T) {
	svc := &Service{permissions: fakeLeavePermissionProvider{
		employeeErr: errors.New("employee lookup failed"),
	}}

	err := svc.authorizeLeaveSubject(context.Background(), 1001, 77, 4102)
	if !errors.Is(err, domain.ErrLeaveAccessDenied) {
		t.Fatalf("error = %v, want ErrLeaveAccessDenied", err)
	}
}

func TestCancelLeaveRequest_DeniesWithoutResolvableEmployee(t *testing.T) {
	svc := &Service{permissions: fakeLeavePermissionProvider{
		employeeErr: errors.New("employee lookup failed"),
	}}

	err := svc.CancelLeaveRequest(context.Background(), 1001, 77, 501)
	if !errors.Is(err, domain.ErrLeaveAccessDenied) {
		t.Fatalf("error = %v, want ErrLeaveAccessDenied", err)
	}
}

func TestCancelLeaveRequest_DeniesZeroCallerUserID(t *testing.T) {
	svc := &Service{permissions: fakeLeavePermissionProvider{employeeID: 4101}}

	err := svc.CancelLeaveRequest(context.Background(), 1001, 0, 501)
	if !errors.Is(err, domain.ErrLeaveAccessDenied) {
		t.Fatalf("error = %v, want ErrLeaveAccessDenied", err)
	}
}

func TestGetLeaveBalances_DeniesWithoutResolvableEmployee(t *testing.T) {
	// repo is deliberately left nil: an early deny must return before ever
	// reaching it, so a nil-pointer panic here would mean the check ran too late.
	svc := &Service{permissions: fakeLeavePermissionProvider{
		employeeErr: errors.New("employee lookup failed"),
	}}

	_, err := svc.GetLeaveBalances(context.Background(), 1001, 77, 4102, 2026)
	if !errors.Is(err, domain.ErrLeaveAccessDenied) {
		t.Fatalf("error = %v, want ErrLeaveAccessDenied", err)
	}
}

func TestGetLeaveBalances_DeniesZeroCallerUserID(t *testing.T) {
	svc := &Service{permissions: fakeLeavePermissionProvider{employeeID: 4101}}

	_, err := svc.GetLeaveBalances(context.Background(), 1001, 0, 4102, 2026)
	if !errors.Is(err, domain.ErrLeaveAccessDenied) {
		t.Fatalf("error = %v, want ErrLeaveAccessDenied", err)
	}
}

func TestApproverMatchesChain_UsesCurrentApprovalLevel(t *testing.T) {
	svc := &Service{permissions: fakeLeavePermissionProvider{roles: map[int64]bool{7: true}}}
	levels := []domain.ApprovalLevelInfo{
		{LevelOrder: 1, ApproverEmployeeID: 4101},
		{LevelOrder: 2, ApproverRoleID: 7},
	}

	matched, err := svc.approverMatchesChain(context.Background(), 1001, 4101, 0, levels)
	if err != nil || !matched {
		t.Fatalf("first approval level matched=%v err=%v", matched, err)
	}
	matched, err = svc.approverMatchesChain(context.Background(), 1001, 4102, 1, levels)
	if err != nil || !matched {
		t.Fatalf("second approval level matched=%v err=%v", matched, err)
	}
	matched, err = svc.approverMatchesChain(context.Background(), 1003, 4103, 0, levels)
	if err != nil {
		t.Fatalf("unmatched approval level: %v", err)
	}
	if matched {
		t.Fatal("unmatched employee should not approve the current level")
	}
}
