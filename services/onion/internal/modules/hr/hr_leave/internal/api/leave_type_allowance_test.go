package api

import (
	"context"
	"testing"

	hrleavev1 "github.com/voxire/voxie/gen/go/hr/leave/v1"
)

func TestUpdateLeaveTypeRequestCarriesTheAllowance(t *testing.T) {
	amount := 21.0
	in := updateLeaveTypeRequestFromProto(5, &hrleavev1.UpdateLeaveTypeBody{AccrualAmount: &amount})

	if in.AccrualAmount == nil || *in.AccrualAmount != 21 {
		t.Fatalf("AccrualAmount = %v, want 21", in.AccrualAmount)
	}
	if err := (&UpdateLeaveTypeHandler{}).Validate(context.Background(), in); err != nil {
		t.Fatalf("an allowance-only update was rejected: %v", err)
	}
}

func TestUpdateLeaveTypeRejectsANegativeAllowance(t *testing.T) {
	amount := -1.0
	in := updateLeaveTypeRequestFromProto(5, &hrleavev1.UpdateLeaveTypeBody{AccrualAmount: &amount})

	if err := (&UpdateLeaveTypeHandler{}).Validate(context.Background(), in); err == nil {
		t.Fatal("a negative allowance was accepted")
	}
}
