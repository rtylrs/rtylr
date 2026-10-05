package api

import (
	"context"
	"net/http"
	"testing"

	hrleavev1 "github.com/voxire/voxie/gen/go/hr/leave/v1"
)

func TestUpdateLeaveTypeHandler_Validate_AllowsNameOnlyPatch(t *testing.T) {
	name := "  Renamed Leave  "
	req := &UpdateLeaveTypeRequest{ID: 42, Name: &name}
	err := (&UpdateLeaveTypeHandler{}).Validate(context.Background(), req)
	if err != nil {
		t.Fatalf("name-only update should be valid: %v", err)
	}
	if req.Name == nil || *req.Name != "Renamed Leave" {
		t.Fatalf("normalized name = %v, want %q", req.Name, "Renamed Leave")
	}
}

func TestUpdateLeaveTypeHandler_Validate_RejectsEmptyAndNegativeFields(t *testing.T) {
	h := &UpdateLeaveTypeHandler{}
	empty := "  "
	if err := h.Validate(context.Background(), &UpdateLeaveTypeRequest{ID: 42, Name: &empty}); oneKitStatus(err) != http.StatusBadRequest {
		t.Fatalf("empty name error = %v, want 400", err)
	}
	negative := int32(-1)
	if err := h.Validate(context.Background(), &UpdateLeaveTypeRequest{ID: 42, MinNoticeDays: &negative}); oneKitStatus(err) != http.StatusBadRequest {
		t.Fatalf("negative notice error = %v, want 400", err)
	}
}

func TestUpdateLeaveTypeHandler_Validate_PreservesFalseAndZeroAsSupplied(t *testing.T) {
	falseValue := false
	zero := int32(0)
	req := &UpdateLeaveTypeRequest{ID: 42, IsPaid: &falseValue, MaxCarryoverDays: &zero}
	if err := (&UpdateLeaveTypeHandler{}).Validate(context.Background(), req); err != nil {
		t.Fatalf("explicit false/zero fields should be accepted: %v", err)
	}
	if req.IsPaid == nil || req.MaxCarryoverDays == nil {
		t.Fatal("presence pointers were lost during validation")
	}
}

func TestUpdateLeaveTypeRequestFromProto_PreservesPatchPresence(t *testing.T) {
	name := "Only name"
	req := updateLeaveTypeRequestFromProto(42, &hrleavev1.UpdateLeaveTypeBody{Name: &name})
	if req.Name == nil || *req.Name != name {
		t.Fatalf("name = %v, want %q", req.Name, name)
	}
	if req.IsPaid != nil || req.RequiresApproval != nil || req.AccrualRule != nil ||
		req.MaxCarryoverDays != nil || req.MinNoticeDays != nil ||
		req.RequiresDocumentationAfterDays != nil || req.IsActive != nil {
		t.Fatalf("omitted generated fields became present: %+v", req)
	}
}
