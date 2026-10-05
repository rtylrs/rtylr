package api

import (
	"context"
	"net/http"
	"testing"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/infrastructure"

	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/application"
	"github.com/voxire/onion/internal/testutil"
)

// --- Authentication gate: no DB needed, these always run ---

func TestListMyPayrollRunsHandler_Serve_DeniedWhenAnonymous(t *testing.T) {
	svc := application.NewService(context.Background(), nil, fakeApprovalChainProvider{}, fakePermissionProvider{employeeID: 4101, allowed: true})
	h := &ListMyPayrollRunsHandler{Service: svc}

	_, err := h.Serve(context.Background(), &ListMyPayrollRunsRequest{Limit: 10})
	if err == nil {
		t.Fatal("expected permission denied error for anonymous caller")
	}
	if got := statusOf(err); got != http.StatusForbidden {
		t.Errorf("status = %d, want %d", got, http.StatusForbidden)
	}
}

func TestGetMyPayrollRunHandler_Serve_DeniedWhenAnonymous(t *testing.T) {
	svc := application.NewService(context.Background(), nil, fakeApprovalChainProvider{}, fakePermissionProvider{employeeID: 4101, allowed: true})
	h := &GetMyPayrollRunHandler{Service: svc}

	ctx := withURLParam(context.Background(), "id", "4301")
	_, err := h.Serve(ctx, &GetMyPayrollRunRequest{})
	if err == nil {
		t.Fatal("expected permission denied error for anonymous caller")
	}
	if got := statusOf(err); got != http.StatusForbidden {
		t.Errorf("status = %d, want %d", got, http.StatusForbidden)
	}
}

// --- Self-service scoping: opt-in DB integration tests ---

func TestListMyPayrollRunsHandler_Serve_ReturnsOnlyCallersRuns(t *testing.T) {
	db := testutil.OpenDB(t)
	svc := application.NewService(context.Background(), moduleinfra.NewRepository(db), fakeApprovalChainProvider{}, fakePermissionProvider{employeeID: 4101, allowed: false})
	h := &ListMyPayrollRunsHandler{Service: svc}

	resp, err := h.Serve(deviceContext(testutil.BusinessID), &ListMyPayrollRunsRequest{Limit: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, r := range resp.Runs {
		if r.Id == 4301 {
			found = true
		}
	}
	if !found {
		t.Error("expected run 4301 in employee 4101's self-service run list")
	}
}

func TestGetMyPayrollRunHandler_Serve_ReturnsOnlyCallersOwnEntry(t *testing.T) {
	db := testutil.OpenDB(t)
	svc := application.NewService(context.Background(), moduleinfra.NewRepository(db), fakeApprovalChainProvider{}, fakePermissionProvider{employeeID: 4101, allowed: false})
	h := &GetMyPayrollRunHandler{Service: svc}

	ctx := withURLParam(deviceContext(testutil.BusinessID), "id", "4301")
	resp, err := h.Serve(ctx, &GetMyPayrollRunRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Run == nil {
		t.Fatal("expected a run payload")
	}
	if len(resp.Run.Entries) != 1 {
		t.Fatalf("expected exactly 1 entry (caller's own), got %d", len(resp.Run.Entries))
	}
	if resp.Run.Entries[0].EmployeeId != 4101 {
		t.Errorf("EmployeeId = %d, want 4101", resp.Run.Entries[0].EmployeeId)
	}
	if resp.Run.Entries[0].NetPay != 1100.0 {
		t.Errorf("NetPay = %v, want 1100 (Alice's own pay, not Bob's)", resp.Run.Entries[0].NetPay)
	}
}

func TestGetMyPayrollRunHandler_Serve_NotFoundWhenCallerHasNoEntry(t *testing.T) {
	db := testutil.OpenDB(t)
	svc := application.NewService(context.Background(), moduleinfra.NewRepository(db), fakeApprovalChainProvider{}, fakePermissionProvider{employeeID: 999999999, allowed: false})
	h := &GetMyPayrollRunHandler{Service: svc}

	ctx := withURLParam(deviceContext(testutil.OwnerID), "id", "4301")
	_, err := h.Serve(ctx, &GetMyPayrollRunRequest{})
	if err == nil {
		t.Fatal("expected not-found error for employee with no entry in this run")
	}
	if got := statusOf(err); got != http.StatusNotFound {
		t.Errorf("status = %d, want %d", got, http.StatusNotFound)
	}
}

// Fixture employee 4102 belongs to business 1001 but carries user id 1002, which
// is the production shape: users.id and businesses.id are independent sequences.
// The real hr_core provider resolves hr_employees by business_id + user_id, so
// it is the only dependency that can catch a handler scoping the lookup by the
// caller's user id; fakePermissionProvider ignores its arguments.
func TestGetMyPayrollRunHandler_Serve_ScopesByCallersBusinessNotUserID(t *testing.T) {
	db := testutil.OpenDB(t)
	svc := application.NewService(context.Background(), moduleinfra.NewRepository(db), fakeApprovalChainProvider{}, hrcore.NewModule(context.Background(), db))
	h := &GetMyPayrollRunHandler{Service: svc}

	ctx := withURLParam(deviceContext(1002), "id", "4301")
	resp, err := h.Serve(ctx, &GetMyPayrollRunRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Run == nil {
		t.Fatal("expected a run payload")
	}
	if len(resp.Run.Entries) != 1 {
		t.Fatalf("expected exactly 1 entry (caller's own), got %d", len(resp.Run.Entries))
	}
	if resp.Run.Entries[0].EmployeeId != 4102 {
		t.Errorf("EmployeeId = %d, want 4102 (Bob's employee row in business 1001)", resp.Run.Entries[0].EmployeeId)
	}
}
