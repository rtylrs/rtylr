package handlers

import (
	"context"
	"testing"
)

// --- CreateHandler.Validate ---

func TestCreateHandler_Validate_RequiresServiceID(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{
		ServiceID:    0,
		CustomerName: "Alice",
		StartTime:    "2026-06-01T09:00:00Z",
		EndTime:      "2026-06-01T10:00:00Z",
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for missing service_id")
	}
}

func TestCreateHandler_Validate_RequiresCustomerName(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{
		ServiceID:    1,
		CustomerName: "",
		StartTime:    "2026-06-01T09:00:00Z",
		EndTime:      "2026-06-01T10:00:00Z",
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for empty customer_name")
	}
}

func TestCreateHandler_Validate_RequiresStartTime(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{
		ServiceID:    1,
		CustomerName: "Bob",
		StartTime:    "",
		EndTime:      "2026-06-01T10:00:00Z",
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for missing start_time")
	}
}

func TestCreateHandler_Validate_RequiresEndTime(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{
		ServiceID:    1,
		CustomerName: "Bob",
		StartTime:    "2026-06-01T09:00:00Z",
		EndTime:      "",
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for missing end_time")
	}
}

func TestCreateHandler_Validate_RejectsInvalidStartTimeFormat(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{
		ServiceID:    1,
		CustomerName: "Bob",
		StartTime:    "June 1st 2026 9am",
		EndTime:      "2026-06-01T10:00:00Z",
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for invalid start_time format")
	}
}

func TestCreateHandler_Validate_RejectsInvalidEndTimeFormat(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{
		ServiceID:    1,
		CustomerName: "Bob",
		StartTime:    "2026-06-01T09:00:00Z",
		EndTime:      "not-a-time",
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for invalid end_time format")
	}
}

func TestCreateHandler_Validate_AcceptsValidRequest(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{
		ServiceID:    1,
		CustomerName: "Alice Smith",
		StartTime:    "2026-06-01T09:00:00Z",
		EndTime:      "2026-06-01T10:00:00Z",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error for valid request: %v", err)
	}
}

func TestCreateHandler_Validate_TrimsWhitespaceFromCustomerName(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{
		ServiceID:    1,
		CustomerName: "   ",
		StartTime:    "2026-06-01T09:00:00Z",
		EndTime:      "2026-06-01T10:00:00Z",
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for whitespace-only customer_name")
	}
}

// --- UpdateHandler.Validate ---

func TestUpdateHandler_Validate_RequiresID(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 0}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestUpdateHandler_Validate_AcceptsValidRequest(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 5}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateHandler_Validate_RejectsInvalidStartTimeFormat(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 5, StartTime: "bad-time"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for invalid start_time format")
	}
}

func TestUpdateHandler_Validate_RejectsInvalidEndTimeFormat(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 5, EndTime: "bad-time"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for invalid end_time format")
	}
}

func TestUpdateHandler_Validate_AllowsEmptyTimes(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 5, StartTime: "", EndTime: ""}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error for empty times: %v", err)
	}
}

// --- DeleteHandler.Validate ---

func TestDeleteHandler_Validate_RequiresID(t *testing.T) {
	h := &DeleteHandler{}
	req := &DeleteRequest{ID: 0}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestDeleteHandler_Validate_AcceptsValidRequest(t *testing.T) {
	h := &DeleteHandler{}
	req := &DeleteRequest{ID: 42}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
