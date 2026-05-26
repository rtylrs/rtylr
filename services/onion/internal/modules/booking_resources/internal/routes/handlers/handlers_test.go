package handlers

import (
	"context"
	"testing"
)

// --- CreateHandler.Validate ---

func TestCreateHandler_Validate_RequiresName(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestCreateHandler_Validate_RejectsWhitespaceOnlyName(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "   "}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for whitespace-only name")
	}
}

func TestCreateHandler_Validate_AcceptsValidRequest(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "Room A", Type: "room", Capacity: 10, IsActive: true}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateHandler_Validate_AcceptsNameOnly(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "Minimal"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateHandler_Validate_TrimsNameBeforeCheck(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "  Valid Name  "}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error after trimming: %v", err)
	}
	if req.Name != "Valid Name" {
		t.Errorf("expected name to be trimmed, got %q", req.Name)
	}
}

// --- UpdateHandler.Validate ---

func TestUpdateHandler_Validate_RequiresID(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 0, Name: "Chair"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestUpdateHandler_Validate_AcceptsValidRequest(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 1, Name: "Chair", Type: "equipment", Capacity: 1}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateHandler_Validate_AcceptsIDOnly(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 3}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
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
	req := &DeleteRequest{ID: 7}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
