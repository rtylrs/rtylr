package api

import (
	"context"
	"testing"
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
