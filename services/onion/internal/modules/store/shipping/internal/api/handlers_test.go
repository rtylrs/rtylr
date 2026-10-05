package api

import (
	"context"
	"testing"
)

func TestCreateHandler_Validate_RequiresName(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "", Price: 5.00}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when name is empty")
	}
}

func TestCreateHandler_Validate_WhitespaceName(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "  \t  ", Price: 0}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for whitespace-only name")
	}
}

func TestCreateHandler_Validate_NegativePrice(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "Standard", Price: -1.0}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for negative price")
	}
}

func TestCreateHandler_Validate_HappyPath(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "Standard Shipping", Price: 9.99}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateHandler_Validate_ZeroPriceAllowed(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "Free Shipping", Price: 0.0}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error for zero price: %v", err)
	}
}

func TestCreateHandler_Validate_WithEstimatedDays(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "Express", Price: 19.99, EstimatedDays: 2}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateHandler_Validate_RequiresID(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 0}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when ID is zero")
	}
}

func TestUpdateHandler_Validate_HappyPath(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 3}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateHandler_Validate_LargeIDPasses(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 99999}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error for large id: %v", err)
	}
}

func TestDeleteHandler_Validate_RequiresID(t *testing.T) {
	h := &DeleteHandler{}
	req := &DeleteRequest{ID: 0}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when ID is zero")
	}
}

func TestDeleteHandler_Validate_HappyPath(t *testing.T) {
	h := &DeleteHandler{}
	req := &DeleteRequest{ID: 5}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeleteHandler_Validate_LargeIDPasses(t *testing.T) {
	h := &DeleteHandler{}
	req := &DeleteRequest{ID: 100000}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error for large id: %v", err)
	}
}

func TestCreateHandler_Validate_LargePriceAllowed(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "International", Price: 9999.99}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error for large price: %v", err)
	}
}

func TestCreateHandler_Validate_ActiveFlag(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "Overnight", Price: 29.99, IsActive: true}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
