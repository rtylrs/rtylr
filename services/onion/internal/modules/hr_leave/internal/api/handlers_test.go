package api

import (
	"context"
	stdErrors "errors"
	"net/http"
	"testing"

	onekithttp "github.com/corezio/onekit/http"
)

// --- CreateLeaveTypeHandler ---

func oneKitStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if stdErrors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func TestCreateLeaveTypeHandler_Validate_MissingName(t *testing.T) {
	h := &CreateLeaveTypeHandler{}
	se := h.Validate(context.Background(), &CreateLeaveTypeRequest{})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for missing name, got %v", se)
	}
}

func TestCreateLeaveTypeHandler_Validate_WhitespaceName(t *testing.T) {
	h := &CreateLeaveTypeHandler{}
	se := h.Validate(context.Background(), &CreateLeaveTypeRequest{Name: "   "})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for whitespace name, got %v", se)
	}
}

func TestCreateLeaveTypeHandler_Validate_ValidName(t *testing.T) {
	h := &CreateLeaveTypeHandler{}
	se := h.Validate(context.Background(), &CreateLeaveTypeRequest{Name: "Annual Leave"})
	if se != nil {
		t.Errorf("expected nil for valid name, got %v", se)
	}
}

// --- CreateLeaveRequestHandler ---

func TestCreateLeaveRequestHandler_Validate_MissingEmployeeID(t *testing.T) {
	h := &CreateLeaveRequestHandler{}
	se := h.Validate(context.Background(), &CreateLeaveRequestRequest{
		LeaveTypeID: 1,
		StartDate:   "2026-06-01",
		EndDate:     "2026-06-05",
		TotalDays:   5,
	})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for missing employee_id, got %v", se)
	}
}

func TestCreateLeaveRequestHandler_Validate_MissingLeaveTypeID(t *testing.T) {
	h := &CreateLeaveRequestHandler{}
	se := h.Validate(context.Background(), &CreateLeaveRequestRequest{
		EmployeeID: 1,
		StartDate:  "2026-06-01",
		EndDate:    "2026-06-05",
		TotalDays:  5,
	})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for missing leave_type_id, got %v", se)
	}
}

func TestCreateLeaveRequestHandler_Validate_MissingDates(t *testing.T) {
	h := &CreateLeaveRequestHandler{}
	se := h.Validate(context.Background(), &CreateLeaveRequestRequest{
		EmployeeID:  1,
		LeaveTypeID: 2,
		TotalDays:   3,
	})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for missing dates, got %v", se)
	}
}

func TestCreateLeaveRequestHandler_Validate_InvalidStartDate(t *testing.T) {
	h := &CreateLeaveRequestHandler{}
	se := h.Validate(context.Background(), &CreateLeaveRequestRequest{
		EmployeeID:  1,
		LeaveTypeID: 2,
		StartDate:   "not-a-date",
		EndDate:     "2026-06-05",
		TotalDays:   3,
	})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid start_date, got %v", se)
	}
}

func TestCreateLeaveRequestHandler_Validate_InvalidEndDate(t *testing.T) {
	h := &CreateLeaveRequestHandler{}
	se := h.Validate(context.Background(), &CreateLeaveRequestRequest{
		EmployeeID:  1,
		LeaveTypeID: 2,
		StartDate:   "2026-06-01",
		EndDate:     "06/05/2026",
		TotalDays:   3,
	})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid end_date, got %v", se)
	}
}

func TestCreateLeaveRequestHandler_Validate_ZeroTotalDays(t *testing.T) {
	h := &CreateLeaveRequestHandler{}
	se := h.Validate(context.Background(), &CreateLeaveRequestRequest{
		EmployeeID:  1,
		LeaveTypeID: 2,
		StartDate:   "2026-06-01",
		EndDate:     "2026-06-01",
		TotalDays:   0,
	})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for zero total_days, got %v", se)
	}
}

func TestCreateLeaveRequestHandler_Validate_ValidRequest(t *testing.T) {
	h := &CreateLeaveRequestHandler{}
	se := h.Validate(context.Background(), &CreateLeaveRequestRequest{
		EmployeeID:  1,
		LeaveTypeID: 2,
		StartDate:   "2026-06-01",
		EndDate:     "2026-06-05",
		TotalDays:   5,
	})
	if se != nil {
		t.Errorf("expected nil for valid request, got %v", se)
	}
}

// --- CancelLeaveRequestHandler ---

func TestCancelLeaveRequestHandler_Validate_MissingRequestID(t *testing.T) {
	h := &CancelLeaveRequestHandler{}
	se := h.Validate(context.Background(), &CancelLeaveRequestRequest{})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for missing request_id, got %v", se)
	}
}

func TestCancelLeaveRequestHandler_Validate_ValidRequest(t *testing.T) {
	h := &CancelLeaveRequestHandler{}
	se := h.Validate(context.Background(), &CancelLeaveRequestRequest{RequestID: 501})
	if se != nil {
		t.Errorf("expected no error for a valid request, got %v", se)
	}
}

// --- ApproveLeaveHandler ---

func TestApproveLeaveHandler_Validate_MissingRequestID(t *testing.T) {
	h := &ApproveLeaveHandler{}
	se := h.Validate(context.Background(), &ApproveLeaveRequest{
		Approved: true,
	})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for missing request_id, got %v", se)
	}
}

func TestApproveLeaveHandler_Validate_RejectWithoutReason(t *testing.T) {
	h := &ApproveLeaveHandler{}
	se := h.Validate(context.Background(), &ApproveLeaveRequest{
		RequestID: 5,
		Approved:  false,
	})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for rejection without reason, got %v", se)
	}
}

func TestApproveLeaveHandler_Validate_RejectWithWhitespaceReason(t *testing.T) {
	h := &ApproveLeaveHandler{}
	se := h.Validate(context.Background(), &ApproveLeaveRequest{
		RequestID:       5,
		Approved:        false,
		RejectionReason: "   ",
	})
	if se == nil || oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400 for whitespace rejection reason, got %v", se)
	}
}

func TestApproveLeaveHandler_Validate_ApproveWithoutReason_Valid(t *testing.T) {
	h := &ApproveLeaveHandler{}
	se := h.Validate(context.Background(), &ApproveLeaveRequest{
		RequestID: 5,
		Approved:  true,
	})
	if se != nil {
		t.Errorf("approve without reason should pass, got %v", se)
	}
}

func TestApproveLeaveHandler_Validate_RejectWithReason_Valid(t *testing.T) {
	h := &ApproveLeaveHandler{}
	se := h.Validate(context.Background(), &ApproveLeaveRequest{
		RequestID:       5,
		Approved:        false,
		RejectionReason: "Insufficient notice",
	})
	if se != nil {
		t.Errorf("reject with reason should pass, got %v", se)
	}
}
