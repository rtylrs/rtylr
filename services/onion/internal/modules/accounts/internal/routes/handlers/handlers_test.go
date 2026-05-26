package handlers

import (
	"context"
	"testing"
)

func TestCreateHandler_Validate_MissingName(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "", Type: "asset"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for missing name")
	}
}

func TestCreateHandler_Validate_MissingType(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "Cash", Type: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for missing type")
	}
}

func TestCreateHandler_Validate_WhitespaceName(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "   ", Type: "asset"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for whitespace-only name")
	}
}

func TestCreateHandler_Validate_Valid(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "Cash", Type: "asset"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateAccountHandler_Validate_MissingName(t *testing.T) {
	h := &UpdateAccountHandler{}
	req := &UpdateAccountRequest{Name: "", Type: "asset"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for missing name")
	}
}

func TestUpdateAccountHandler_Validate_MissingType(t *testing.T) {
	h := &UpdateAccountHandler{}
	req := &UpdateAccountRequest{Name: "Cash", Type: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for missing type")
	}
}

func TestUpdateAccountHandler_Validate_Valid(t *testing.T) {
	h := &UpdateAccountHandler{}
	req := &UpdateAccountRequest{Name: "Cash", Type: "asset"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
