package handlers

import (
	"context"
	"testing"
)

// --- CreateHandler.Validate ---

func TestCreateHandler_Validate_RequiresCustomerName(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{CustomerName: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for empty customer_name")
	}
}

func TestCreateHandler_Validate_RejectsWhitespaceOnlyName(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{CustomerName: "   "}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for whitespace-only customer_name")
	}
}

func TestCreateHandler_Validate_AcceptsValidRequest(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{
		CustomerName:  "Jane Doe",
		CustomerEmail: "jane@example.com",
		ServiceID:     1,
		PreferredDate: "2026-07-01",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateHandler_Validate_AcceptsNameOnly(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{CustomerName: "Bob"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateHandler_Validate_TrimsCustomerName(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{CustomerName: "  Alice  "}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.CustomerName != "Alice" {
		t.Errorf("expected trimmed name, got %q", req.CustomerName)
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
	req := &UpdateRequest{
		ID:           1,
		CustomerName: "Tom",
		Status:       "pending",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateHandler_Validate_AcceptsIDOnly(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 5}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateHandler_Validate_AcceptsStatusChange(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 10, Status: "confirmed"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
