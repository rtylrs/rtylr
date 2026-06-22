package handlers

import (
	"context"
	"net/http"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
)

// ─── AssignLocationHandler.Validate ─────────────────────────────────────────

func TestAssign_Validate_AcceptsValidRequest(t *testing.T) {
	h := &AssignLocationHandler{}
	req := &AssignLocationRequest{UserID: 1, LocationID: 2}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("valid request should pass: %v", err.Message)
	}
}

func TestAssign_Validate_RejectsZeroUserID(t *testing.T) {
	h := &AssignLocationHandler{}
	if err := h.Validate(context.Background(), &AssignLocationRequest{LocationID: 1}); err == nil {
		t.Error("expected error for zero user_id")
	}
}

func TestAssign_Validate_RejectsZeroLocationID(t *testing.T) {
	h := &AssignLocationHandler{}
	if err := h.Validate(context.Background(), &AssignLocationRequest{UserID: 1}); err == nil {
		t.Error("expected error for zero location_id")
	}
}

func TestAssign_Validate_RejectsBothZero(t *testing.T) {
	h := &AssignLocationHandler{}
	if err := h.Validate(context.Background(), &AssignLocationRequest{}); err == nil {
		t.Error("expected error for both zero")
	}
}

func TestAssign_Validate_RejectsNegativeIDs(t *testing.T) {
	h := &AssignLocationHandler{}
	if err := h.Validate(context.Background(), &AssignLocationRequest{UserID: -1, LocationID: 2}); err == nil {
		t.Error("expected error for negative user_id")
	}
	if err := h.Validate(context.Background(), &AssignLocationRequest{UserID: 1, LocationID: -2}); err == nil {
		t.Error("expected error for negative location_id")
	}
}

func TestAssign_Validate_PermissionsDefaultFalse(t *testing.T) {
	req := &AssignLocationRequest{}
	if req.CanManagePOSUsers || req.CanManageTerminals || req.CanViewReports {
		t.Error("all permission flags should default to false")
	}
}

func TestAssign_Validate_PermissionsCanBeTrue(t *testing.T) {
	h := &AssignLocationHandler{}
	req := &AssignLocationRequest{
		UserID:             5,
		LocationID:         10,
		CanManagePOSUsers:  true,
		CanManageTerminals: true,
		CanViewReports:     true,
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("all-true permissions should still be valid: %v", err.Message)
	}
}

func TestAssign_Validate_AcceptsLegacyCamelCaseRequest(t *testing.T) {
	h := &AssignLocationHandler{}
	canManagePOSUsers := true
	canManageTerminals := true
	canViewReports := true
	req := &AssignLocationRequest{
		UserIDCamel:             5,
		LocationIDCamel:         10,
		CanManagePOSUsersCamel:  &canManagePOSUsers,
		CanManageTerminalsCamel: &canManageTerminals,
		CanViewReportsCamel:     &canViewReports,
	}

	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("camelCase request should pass: %v", err.Message)
	}
	if req.UserID != 5 || req.LocationID != 10 {
		t.Fatalf("aliases not normalized: user=%d location=%d", req.UserID, req.LocationID)
	}
	if !req.CanManagePOSUsers || !req.CanManageTerminals || !req.CanViewReports {
		t.Fatal("camelCase management flags were not normalized")
	}
}

// ─── RemoveLocationHandler.Validate ─────────────────────────────────────────

func TestRemove_Validate_AcceptsNonZeroID(t *testing.T) {
	h := &RemoveLocationHandler{}
	if err := h.Validate(context.Background(), &RemoveLocationRequest{ID: 7}); err != nil {
		t.Errorf("non-zero ID should pass: %v", err.Message)
	}
}

func TestRemove_Validate_RejectsZeroID(t *testing.T) {
	h := &RemoveLocationHandler{}
	if err := h.Validate(context.Background(), &RemoveLocationRequest{ID: 0}); err == nil {
		t.Error("expected error for zero ID")
	}
}

func TestRemove_Validate_RejectsNegativeID(t *testing.T) {
	h := &RemoveLocationHandler{}
	if err := h.Validate(context.Background(), &RemoveLocationRequest{ID: -1}); err == nil {
		t.Error("expected error for negative ID")
	}
}

func TestAssign_Serve_RejectsNonAdminUser(t *testing.T) {
	h := &AssignLocationHandler{}
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{UserID: 2, BusinessID: 1, Role: "store_manager"})

	_, err := h.Serve(ctx, &AssignLocationRequest{UserID: 3, LocationID: 4})
	if err == nil || err.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %+v", err)
	}
}

func TestRemove_Serve_RejectsNonAdminUser(t *testing.T) {
	h := &RemoveLocationHandler{}
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{UserID: 2, BusinessID: 1, Role: "staff"})

	_, err := h.Serve(ctx, &RemoveLocationRequest{ID: 4})
	if err == nil || err.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %+v", err)
	}
}

func TestIsOwnerOrAdmin_UsesRoleNotIDCollision(t *testing.T) {
	if isOwnerOrAdmin(&authPkg.AuthUser{UserID: 1, BusinessID: 1, Role: "staff"}) {
		t.Fatal("staff user must not be privileged just because user_id matches business_id")
	}
	if !isOwnerOrAdmin(&authPkg.AuthUser{UserID: 2, BusinessID: 1, Role: "admin"}) {
		t.Fatal("admin should be privileged")
	}
	if !isOwnerOrAdmin(&authPkg.AuthUser{UserID: 2, BusinessID: 1, Role: "console_admin"}) {
		t.Fatal("console admin should be privileged")
	}
	if !isOwnerOrAdmin(&authPkg.AuthUser{UserID: 1, BusinessID: 1, Role: "owner"}) {
		t.Fatal("owner should be privileged")
	}
}

func TestHasBusinessWidePOSAccess_RoleRules(t *testing.T) {
	if !hasBusinessWidePOSAccess(context.Background(), &authPkg.AuthUser{UserID: 2, BusinessID: 1, Role: "owner"}) {
		t.Fatal("owner role should be privileged")
	}
	if !hasBusinessWidePOSAccess(context.Background(), &authPkg.AuthUser{UserID: 2, BusinessID: 1, Role: "admin"}) {
		t.Fatal("admin role should be privileged")
	}
	if !hasBusinessWidePOSAccess(context.Background(), &authPkg.AuthUser{UserID: 2, BusinessID: 1, Role: "console_admin"}) {
		t.Fatal("console admin role should be privileged")
	}
	if hasBusinessWidePOSAccess(context.Background(), &authPkg.AuthUser{UserID: 2, BusinessID: 1, Role: "store_manager"}) {
		t.Fatal("store manager should not be privileged without matching business owner row")
	}
	if hasBusinessWidePOSAccess(context.Background(), nil) {
		t.Fatal("nil auth user should not be privileged")
	}
}

// ─── Handler struct creation ──────────────────────────────────────────────────

func TestHandlers_CanBeInstantiated(_ *testing.T) {
	_ = &GetMyLocationsHandler{}
	_ = &GetLocationRosterHandler{}
	_ = &GetMemberLocationsHandler{}
	_ = &AssignLocationHandler{}
	_ = &RemoveLocationHandler{}
}

// ─── Routes helpers ───────────────────────────────────────────────────────────

func TestRoutes_NonEmpty(t *testing.T) {
	for name, fn := range map[string]func() []string{
		"MyLocations":     MyLocationsRoutes,
		"LocationRoster":  LocationRosterRoutes,
		"MemberLocations": MemberLocationsRoutes,
		"Assign":          AssignRoutes,
		"Remove":          RemoveRoutes,
	} {
		if len(fn()) == 0 {
			t.Errorf("%s: expected at least one route", name)
		}
	}
}

// ─── Response struct defaults ─────────────────────────────────────────────────

func TestGetMyLocationsResponse_DefaultEmptyLocations(t *testing.T) {
	resp := GetMyLocationsResponse{}
	if resp.Locations != nil {
		t.Error("expected nil locations by default")
	}
}

func TestAssignLocationResponse_DefaultZeroID(t *testing.T) {
	resp := AssignLocationResponse{}
	if resp.ID != 0 {
		t.Error("expected zero ID by default")
	}
}

func TestRemoveLocationResponse_DefaultFalse(t *testing.T) {
	resp := RemoveLocationResponse{}
	if resp.Success {
		t.Error("expected false Success by default")
	}
}
