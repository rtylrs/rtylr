package handlers

import (
	"context"
	"testing"
)

func TestCreateVariantHandler_Validate_RequiresProductID(t *testing.T) {
	h := &CreateVariantHandler{}
	req := &CreateVariantRequest{ProductID: 0, Name: "Red / L"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when ProductID is zero")
	}
}

func TestCreateVariantHandler_Validate_RequiresName(t *testing.T) {
	h := &CreateVariantHandler{}
	req := &CreateVariantRequest{ProductID: 5, Name: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when name is empty")
	}
}

func TestCreateVariantHandler_Validate_WhitespaceName(t *testing.T) {
	h := &CreateVariantHandler{}
	req := &CreateVariantRequest{ProductID: 5, Name: "  \t  "}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for whitespace-only name")
	}
}

func TestCreateVariantHandler_Validate_HappyPath(t *testing.T) {
	h := &CreateVariantHandler{}
	req := &CreateVariantRequest{ProductID: 5, Name: "Red / Large"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateVariantHandler_Validate_WithOptionalFields(t *testing.T) {
	h := &CreateVariantHandler{}
	req := &CreateVariantRequest{
		ProductID: 10,
		Name:      "Blue / Small",
		SKU:       "SKU-001",
		SalePrice: 29.99,
		IsActive:  true,
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error with optional fields: %v", err)
	}
}

func TestCreateVariantHandler_Validate_BothEmpty(t *testing.T) {
	h := &CreateVariantHandler{}
	req := &CreateVariantRequest{}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when all fields are empty")
	}
}

func TestUpdateVariantHandler_Validate_RequiresID(t *testing.T) {
	h := &UpdateVariantHandler{}
	req := &UpdateVariantRequest{ID: 0, Name: "Blue / M"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when ID is zero")
	}
}

func TestUpdateVariantHandler_Validate_RequiresName(t *testing.T) {
	h := &UpdateVariantHandler{}
	req := &UpdateVariantRequest{ID: 3, Name: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when name is empty")
	}
}

func TestUpdateVariantHandler_Validate_HappyPath(t *testing.T) {
	h := &UpdateVariantHandler{}
	req := &UpdateVariantRequest{ID: 5, Name: "Green / XL"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateVariantHandler_Validate_WhitespaceName(t *testing.T) {
	h := &UpdateVariantHandler{}
	req := &UpdateVariantRequest{ID: 1, Name: "   "}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for whitespace name")
	}
}

func TestDeleteVariantHandler_Validate_RequiresID(t *testing.T) {
	h := &DeleteVariantHandler{}
	req := &DeleteVariantRequest{ID: 0}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when ID is zero")
	}
}

func TestDeleteVariantHandler_Validate_HappyPath(t *testing.T) {
	h := &DeleteVariantHandler{}
	req := &DeleteVariantRequest{ID: 7}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeleteVariantHandler_Validate_LargeIDPasses(t *testing.T) {
	h := &DeleteVariantHandler{}
	req := &DeleteVariantRequest{ID: 99999}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error for large id: %v", err)
	}
}

func TestUpdateVariantHandler_Validate_LargeIDPasses(t *testing.T) {
	h := &UpdateVariantHandler{}
	req := &UpdateVariantRequest{ID: 10000, Name: "White / S"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error for large id: %v", err)
	}
}
