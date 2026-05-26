package handlers

import (
	"context"
	"testing"
)

func TestCreateCannedHandler_Validate_RequiresName(t *testing.T) {
	h := &CreateCannedHandler{}
	req := &CreateCannedRequest{Name: "", Body: "Some body"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when name is empty")
	}
}

func TestCreateCannedHandler_Validate_RequiresBody(t *testing.T) {
	h := &CreateCannedHandler{}
	req := &CreateCannedRequest{Name: "Greeting", Body: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when body is empty")
	}
}

func TestCreateCannedHandler_Validate_HappyPath(t *testing.T) {
	h := &CreateCannedHandler{}
	req := &CreateCannedRequest{Name: "Greeting", Body: "Hello, how can I help?"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateCannedHandler_Validate_WhitespaceName(t *testing.T) {
	h := &CreateCannedHandler{}
	req := &CreateCannedRequest{Name: "   ", Body: "Some body"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for whitespace-only name")
	}
}

func TestCreateCannedHandler_Validate_WhitespaceBody(t *testing.T) {
	h := &CreateCannedHandler{}
	req := &CreateCannedRequest{Name: "Hello", Body: "  \t  "}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for whitespace-only body")
	}
}

func TestCreateCannedHandler_Validate_WithOptionalFields(t *testing.T) {
	h := &CreateCannedHandler{}
	req := &CreateCannedRequest{
		Name:     "Billing",
		Body:     "Your invoice is attached.",
		Category: "billing",
		Shortcut: "/billing",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error with optional fields: %v", err)
	}
}

func TestUpdateCannedHandler_Validate_RequiresID(t *testing.T) {
	h := &UpdateCannedHandler{}
	req := &UpdateCannedRequest{ID: 0, Name: "Updated"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when ID is zero")
	}
}

func TestUpdateCannedHandler_Validate_HappyPath(t *testing.T) {
	h := &UpdateCannedHandler{}
	req := &UpdateCannedRequest{ID: 5, Name: "Updated"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeleteCannedHandler_Validate_RequiresID(t *testing.T) {
	h := &DeleteCannedHandler{}
	req := &DeleteCannedRequest{ID: 0}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when ID is zero")
	}
}

func TestDeleteCannedHandler_Validate_HappyPath(t *testing.T) {
	h := &DeleteCannedHandler{}
	req := &DeleteCannedRequest{ID: 7}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateCannedHandler_Validate_BothEmpty(t *testing.T) {
	h := &CreateCannedHandler{}
	req := &CreateCannedRequest{}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when both name and body are empty")
	}
}

func TestUpdateCannedHandler_Validate_NonZeroIDPasses(t *testing.T) {
	h := &UpdateCannedHandler{}
	req := &UpdateCannedRequest{ID: 1}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error for non-zero id: %v", err)
	}
}

func TestDeleteCannedHandler_Validate_LargeIDPasses(t *testing.T) {
	h := &DeleteCannedHandler{}
	req := &DeleteCannedRequest{ID: 999999}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error for large id: %v", err)
	}
}

func TestUpdateCannedHandler_Validate_LargeIDPasses(t *testing.T) {
	h := &UpdateCannedHandler{}
	req := &UpdateCannedRequest{ID: 10000}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error for large id: %v", err)
	}
}
