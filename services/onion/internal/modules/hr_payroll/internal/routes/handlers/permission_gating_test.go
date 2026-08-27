package handlers

import (
	"context"
	stdErrors "errors"
	"net/http"
	"strconv"
	"testing"

	onekithttp "github.com/corezio/onekit/http"
	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/application"
	"github.com/voxire/onion/internal/modules/hr_payroll/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_payroll/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

type fakePermissionProvider struct {
	employeeID int64
	allowed    bool
}

func (f fakePermissionProvider) GetEmployeeIDByUserID(context.Context, int64, int64) (int64, error) {
	return f.employeeID, nil
}

func (f fakePermissionProvider) HasPermission(context.Context, int64, int64, string) (bool, error) {
	return f.allowed, nil
}

func (f fakePermissionProvider) HasRole(context.Context, int64, int64, int64) (bool, error) {
	return f.allowed, nil
}

type fakeApprovalChainProvider struct{}

func (fakeApprovalChainProvider) GetApprovalChain(context.Context, int64, string) (*domain.ApprovalChainInfo, error) {
	return &domain.ApprovalChainInfo{TotalLevels: 1}, nil
}

func deviceContext(userID int64) context.Context {
	device := &authPkg.Device{User: &authPkg.AuthUser{
		UserID:     userID,
		BusinessID: testutil.BusinessID,
		EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
			AppID:     authorization.AppHR,
			Licensed:  true,
			HasAccess: true,
			Permissions: []string{
				"hr:payroll:view",
				"hr:payroll:manage",
				"hr:payroll:execute",
			},
		}}},
	}}
	return device.Ctx(context.Background())
}

func withURLParam(ctx context.Context, key, value string) context.Context {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return context.WithValue(ctx, chi.RouteCtxKey, rctx)
}

func statusOf(err error) int {
	var oneKitErr *onekithttp.Error
	if stdErrors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

// --- Denied path: no DB access should ever happen, so these run without DB_HOST ---

func TestTriggerPayrollRunHandler_Serve_DeniedWithoutManagePayroll(t *testing.T) {
	svc := application.NewService(context.Background(), nil, fakeApprovalChainProvider{}, fakePermissionProvider{employeeID: 4102, allowed: false})
	h := &TriggerPayrollRunHandler{Service: svc}

	_, err := h.Serve(deviceContext(1001), &TriggerPayrollRunRequest{PeriodStart: "2026-04-01", PeriodEnd: "2026-04-30"})
	if err == nil {
		t.Fatal("expected permission denied error")
	}
	if got := statusOf(err); got != http.StatusForbidden {
		t.Errorf("status = %d, want %d", got, http.StatusForbidden)
	}
}

func TestApprovePayrollRunHandler_Serve_DeniedWithoutManagePayroll(t *testing.T) {
	svc := application.NewService(context.Background(), nil, fakeApprovalChainProvider{}, fakePermissionProvider{employeeID: 4102, allowed: false})
	h := &ApprovePayrollRunHandler{Service: svc}

	ctx := withURLParam(deviceContext(1001), "id", "4301")
	_, err := h.Serve(ctx, &ApprovePayrollRunRequest{})
	if err == nil {
		t.Fatal("expected permission denied error")
	}
	if got := statusOf(err); got != http.StatusForbidden {
		t.Errorf("status = %d, want %d", got, http.StatusForbidden)
	}
}

func TestGetPayrollRunHandler_Serve_DeniedWithoutManagePayroll(t *testing.T) {
	svc := application.NewService(context.Background(), nil, fakeApprovalChainProvider{}, fakePermissionProvider{employeeID: 4102, allowed: false})
	h := &GetPayrollRunHandler{Service: svc}

	ctx := withURLParam(deviceContext(1001), "id", "4301")
	_, err := h.Serve(ctx, &GetPayrollRunRequest{})
	if err == nil {
		t.Fatal("expected permission denied error")
	}
	if got := statusOf(err); got != http.StatusForbidden {
		t.Errorf("status = %d, want %d", got, http.StatusForbidden)
	}
}

func TestListPayrollRunsHandler_Serve_DeniedWithoutManagePayroll(t *testing.T) {
	svc := application.NewService(context.Background(), nil, fakeApprovalChainProvider{}, fakePermissionProvider{employeeID: 4102, allowed: false})
	h := &ListPayrollRunsHandler{Service: svc}

	_, err := h.Serve(deviceContext(1001), &ListPayrollRunsRequest{Limit: 10})
	if err == nil {
		t.Fatal("expected permission denied error")
	}
	if got := statusOf(err); got != http.StatusForbidden {
		t.Errorf("status = %d, want %d", got, http.StatusForbidden)
	}
}

// --- Allowed path: real repo access, so these are opt-in DB integration tests ---

func TestGetPayrollRunHandler_Serve_SucceedsWithManagePayroll(t *testing.T) {
	db := testutil.OpenDB(t)
	svc := application.NewService(context.Background(), db, fakeApprovalChainProvider{}, fakePermissionProvider{employeeID: 4101, allowed: true})
	h := &GetPayrollRunHandler{Service: svc}

	ctx := withURLParam(deviceContext(testutil.BusinessID), "id", "4301")
	resp, err := h.Serve(ctx, &GetPayrollRunRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Run == nil || resp.Run.Id != 4301 {
		t.Errorf("expected run 4301, got %+v", resp.Run)
	}
}

func TestListPayrollRunsHandler_Serve_SucceedsWithManagePayroll(t *testing.T) {
	db := testutil.OpenDB(t)
	svc := application.NewService(context.Background(), db, fakeApprovalChainProvider{}, fakePermissionProvider{employeeID: 4101, allowed: true})
	h := &ListPayrollRunsHandler{Service: svc}

	resp, err := h.Serve(deviceContext(testutil.OwnerID), &ListPayrollRunsRequest{Limit: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total < 1 {
		t.Errorf("expected at least 1 run, got total=%d", resp.Total)
	}
}

func TestApprovePayrollRunHandler_Serve_SucceedsWithManagePayroll(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	run, _, err := repo.CreatePayrollRunWithEntries(ctx, testutil.OwnerID, domain.PayrollRunRecord{
		PeriodStart: "2026-11-01",
		PeriodEnd:   "2026-11-30",
		Status:      domain.PayrollStatusDraft,
	}, []domain.PayrollEntryRecord{})
	if err != nil {
		t.Fatalf("CreatePayrollRunWithEntries: %v", err)
	}

	svc := application.NewService(ctx, db, fakeApprovalChainProvider{}, fakePermissionProvider{employeeID: 4101, allowed: true})
	h := &ApprovePayrollRunHandler{Service: svc}

	reqCtx := withURLParam(deviceContext(testutil.OwnerID), "id", strconv.FormatInt(run.ID, 10))
	resp, err := h.Serve(reqCtx, &ApprovePayrollRunRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
}
