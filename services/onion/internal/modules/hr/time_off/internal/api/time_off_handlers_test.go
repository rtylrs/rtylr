package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/time_off/internal/application"
	"github.com/voxire/onion/internal/modules/hr/time_off/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/tenant"
)

func TestSubmitTimeOffHandler_Validate(t *testing.T) {
	h := &SubmitTimeOffHandler{}

	// Missing staff ID
	req := &SubmitTimeOffRequest{
		StartDate: "2026-10-15",
		EndDate:   "2026-10-17",
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for missing staff ID")
	}

	// Missing dates
	req = &SubmitTimeOffRequest{
		StaffID: 123,
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for missing dates")
	}

	// End before start
	req = &SubmitTimeOffRequest{
		StaffID:   123,
		StartDate: "2026-10-20",
		EndDate:   "2026-10-15",
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error when end date is before start date")
	}

	// Malformed dates
	req = &SubmitTimeOffRequest{
		StaffID:   123,
		StartDate: "15/10/2026",
		EndDate:   "2026-10-20",
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for non YYYY-MM-DD date")
	}

	// Valid with camelCase
	req = &SubmitTimeOffRequest{
		StaffID:   123,
		StartDate: "2026-10-15",
		EndDate:   "2026-10-20",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error for valid camelCase request: %v", err)
	}

	// Valid with snake_case
	req = &SubmitTimeOffRequest{
		StaffIDSnake:   123,
		StartDateSnake: "2026-10-15",
		EndDateSnake:   "2026-10-20",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error for valid snake_case request: %v", err)
	}
}

func TestApproveTimeOffHandler_Validate(t *testing.T) {
	h := &ApproveTimeOffHandler{}

	// Missing ID
	req := &ApproveTimeOffRequest{
		Status: "approved",
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for missing ID")
	}

	// Invalid status
	req = &ApproveTimeOffRequest{
		ID:     1,
		Status: "invalid_status",
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for invalid status")
	}

	// Valid approved
	req = &ApproveTimeOffRequest{
		ID:     1,
		Status: "approved",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Valid rejected
	req = &ApproveTimeOffRequest{
		ID:     1,
		Status: "rejected",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

type fakeTimeOffRepo struct {
	all          []domain.TimeOffRequest
	own          []domain.TimeOffRequest
	owns         bool
	listedForAll bool
	created      bool
}

func (f *fakeTimeOffRepo) ListTimeOffRequests(context.Context, int64) ([]domain.TimeOffRequest, error) {
	f.listedForAll = true
	return f.all, nil
}

func (f *fakeTimeOffRepo) ListTimeOffRequestsForUser(context.Context, int64, int64) ([]domain.TimeOffRequest, error) {
	return f.own, nil
}

func (f *fakeTimeOffRepo) StaffBelongsToUser(context.Context, int64, int64, int64) (bool, error) {
	return f.owns, nil
}

func (f *fakeTimeOffRepo) CreateTimeOffRequest(context.Context, domain.TimeOffInput) (int64, error) {
	f.created = true
	return 1, nil
}

func (f *fakeTimeOffRepo) ApproveTimeOffRequest(context.Context, int64, domain.TimeOffApprovalInput) error {
	return nil
}

type fixedScope struct{}

func (fixedScope) ActiveBusinessOwner(_ context.Context, businessID int64) (int64, error) {
	return businessID, nil
}

func (fixedScope) UserBusiness(_ context.Context, businessID, _ int64) (int64, error) {
	return businessID, nil
}

func hrCtx(t *testing.T, permissions ...string) context.Context {
	t.Helper()
	t.Cleanup(tenant.SetScopeResolver(fixedScope{}))
	return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppHR,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
			}},
		},
	})
}

func TestListTimeOff_EmployeesOnlySeeTheirOwnRequests(t *testing.T) {
	repo := &fakeTimeOffRepo{
		all: []domain.TimeOffRequest{{ID: 1, StaffID: 1}, {ID: 2, StaffID: 2}},
		own: []domain.TimeOffRequest{{ID: 2, StaffID: 2}},
	}
	h := &ListTimeOffHandler{Service: application.NewService(context.Background(), repo)}

	got, err := h.Serve(hrCtx(t, "hr:time_off:request"), &ListTimeOffRequest{})
	if err != nil {
		t.Fatalf("Serve: %v", err)
	}
	if len(got) != 1 || got[0].ID != 2 || repo.listedForAll {
		t.Fatalf("employee saw %v (listedForAll=%v), want only their own request", got, repo.listedForAll)
	}

	got, err = h.Serve(hrCtx(t, "hr:time_off:approve"), &ListTimeOffRequest{})
	if err != nil || len(got) != 2 || !repo.listedForAll {
		t.Fatalf("approver got %v, %v; want the whole business", got, err)
	}
}

func TestSubmitTimeOff_EmployeesCannotFileForColleagues(t *testing.T) {
	req := &SubmitTimeOffRequest{StaffID: 5, Type: "sick", StartDate: "2026-10-20", EndDate: "2026-10-21"}

	repo := &fakeTimeOffRepo{owns: false}
	h := &SubmitTimeOffHandler{Service: application.NewService(context.Background(), repo)}
	if _, err := h.Serve(hrCtx(t, "hr:time_off:request"), req); err == nil || repo.created {
		t.Fatalf("employee filing for someone else: err=%v created=%v, want denied", err, repo.created)
	}

	repo = &fakeTimeOffRepo{owns: true}
	h = &SubmitTimeOffHandler{Service: application.NewService(context.Background(), repo)}
	if _, err := h.Serve(hrCtx(t, "hr:time_off:request"), req); err != nil || !repo.created {
		t.Fatalf("employee filing for themself: err=%v created=%v, want success", err, repo.created)
	}

	repo = &fakeTimeOffRepo{owns: false}
	h = &SubmitTimeOffHandler{Service: application.NewService(context.Background(), repo)}
	if _, err := h.Serve(hrCtx(t, "hr:time_off:approve"), req); err != nil || !repo.created {
		t.Fatalf("approver filing for a colleague: err=%v created=%v, want success", err, repo.created)
	}
}
