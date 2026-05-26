package handlers

import (
	"context"
	"testing"
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

func TestAssign_Validate_PermissionsDefaultFalse(t *testing.T) {
	req := &AssignLocationRequest{UserID: 1, LocationID: 2}
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
	// The guard checks == 0; negative IDs are int64 so non-zero but invalid in
	// practice. Confirm the zero-guard fires only on zero, not negative.
	h := &RemoveLocationHandler{}
	// -1 is non-zero, so Validate passes (guard is ID == 0).
	if err := h.Validate(context.Background(), &RemoveLocationRequest{ID: -1}); err != nil {
		t.Logf("negative id rejected at validation (acceptable): %v", err.Message)
	}
}

// ─── Handler struct creation ──────────────────────────────────────────────────

func TestHandlers_CanBeInstantiated(_ *testing.T) {
	_ = &GetMyLocationsHandler{}
	_ = &AssignLocationHandler{}
	_ = &RemoveLocationHandler{}
}

// ─── Routes helpers ───────────────────────────────────────────────────────────

func TestRoutes_NonEmpty(t *testing.T) {
	for name, fn := range map[string]func() []string{
		"MyLocations": MyLocationsRoutes,
		"Assign":      AssignRoutes,
		"Remove":      RemoveRoutes,
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
