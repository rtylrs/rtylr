package handlers

import (
	"context"
	"testing"
)

// --- CreateHandler.Validate ---

func TestCreateHandler_Validate_RequiresName(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "", DurationMinutes: 30, Price: 0}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestCreateHandler_Validate_RejectsWhitespaceOnlyName(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "   ", DurationMinutes: 30, Price: 0}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for whitespace-only name")
	}
}

func TestCreateHandler_Validate_RequiresDurationMinutes(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "Haircut", DurationMinutes: 0, Price: 10}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for zero duration_minutes")
	}
}

func TestCreateHandler_Validate_RejectsNegativeDuration(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "Haircut", DurationMinutes: -5, Price: 10}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for negative duration_minutes")
	}
}

func TestCreateHandler_Validate_RejectsNegativePrice(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "Haircut", DurationMinutes: 30, Price: -1}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for negative price")
	}
}

func TestCreateHandler_Validate_AcceptsZeroPrice(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "Free Consultation", DurationMinutes: 15, Price: 0}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error for zero price: %v", err)
	}
}

func TestCreateHandler_Validate_AcceptsValidRequest(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "Deep Tissue Massage", DurationMinutes: 60, Price: 100.0, IsActive: true}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateHandler_Validate_TrimsNameBeforeCheck(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "  Facial  ", DurationMinutes: 45, Price: 50}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Name != "Facial" {
		t.Errorf("expected trimmed name, got %q", req.Name)
	}
}

// --- UpdateHandler.Validate ---

func TestUpdateHandler_Validate_RequiresID(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 0, Name: "Manicure", DurationMinutes: 30, Price: 25}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestUpdateHandler_Validate_RequiresName(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 1, Name: "", DurationMinutes: 30, Price: 25}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for empty name on update")
	}
}

func TestUpdateHandler_Validate_RequiresDurationMinutes(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 1, Name: "Service", DurationMinutes: 0, Price: 25}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for zero duration_minutes on update")
	}
}

func TestUpdateHandler_Validate_RejectsNegativePrice(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 1, Name: "Service", DurationMinutes: 30, Price: -5}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for negative price on update")
	}
}

func TestUpdateHandler_Validate_AcceptsValidRequest(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 2, Name: "Pedicure", DurationMinutes: 45, Price: 35.0}
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
	req := &DeleteRequest{ID: 9}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
