package api

import (
	"context"
	"testing"

	hrovertimev1 "github.com/voxire/voxie/gen/go/hr/overtime/v1"
)

func TestOvertimeListRequestFromProto_ForwardsAndNormalizesFilters(t *testing.T) {
	employeeID := int64(42)
	status := " PENDING "
	got, err := overtimeListRequestFromProto(&hrovertimev1.ListOvertimeRequestsRequest{
		EmployeeId: &employeeID,
		Status:     &status,
	})
	if err != nil {
		t.Fatalf("mapping filters: %v", err)
	}
	if got.EmployeeID == nil || *got.EmployeeID != 42 {
		t.Fatalf("employee filter = %#v, want 42", got.EmployeeID)
	}
	if got.Status == nil || *got.Status != "pending" {
		t.Fatalf("status filter = %#v, want pending", got.Status)
	}
}

func TestOvertimeApprovalRequestFromProto_PreservesRejectionReason(t *testing.T) {
	reason := "Insufficient coverage"
	got := overtimeApprovalRequestFromProto(&hrovertimev1.ApproveOvertimeRequest{
		OvertimeRequestId: 17,
		Body: &hrovertimev1.ApproveOvertimeBody{
			Approved:        false,
			RejectionReason: &reason,
		},
	})
	if got.RequestID != 17 || got.Approved || got.RejectionReason != reason {
		t.Fatalf("mapped approval request = %+v, want id=17 rejected reason=%q", got, reason)
	}
}

func TestApproveOvertimeHandler_RejectsBlankReason(t *testing.T) {
	err := (&ApproveOvertimeHandler{}).Validate(context.Background(), &ApproveOvertimeRequestReq{
		RequestID:       17,
		Approved:        false,
		RejectionReason: " \t ",
	})
	if err == nil {
		t.Fatal("blank rejection reason was accepted")
	}
}
