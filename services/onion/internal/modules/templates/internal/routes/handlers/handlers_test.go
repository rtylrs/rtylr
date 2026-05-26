package handlers

import (
	"context"
	"testing"
)

func TestCreateTemplateHandler_Validate_RequiresName(t *testing.T) {
	h := &CreateTemplateHandler{}
	req := &CreateTemplateRequest{Name: "", Type: "email", Body: "Hello {{name}}"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when name is empty")
	}
}

func TestCreateTemplateHandler_Validate_RequiresType(t *testing.T) {
	h := &CreateTemplateHandler{}
	req := &CreateTemplateRequest{Name: "Welcome", Type: "", Body: "Hello {{name}}"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when type is empty")
	}
}

func TestCreateTemplateHandler_Validate_RequiresBody(t *testing.T) {
	h := &CreateTemplateHandler{}
	req := &CreateTemplateRequest{Name: "Welcome", Type: "email", Body: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when body is empty")
	}
}

func TestCreateTemplateHandler_Validate_HappyPath(t *testing.T) {
	h := &CreateTemplateHandler{}
	req := &CreateTemplateRequest{
		Name: "Welcome",
		Type: "email",
		Body: "Hello {{name}}",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateTemplateHandler_Validate_WhitespaceName(t *testing.T) {
	h := &CreateTemplateHandler{}
	req := &CreateTemplateRequest{Name: "   ", Type: "sms", Body: "Hi {{name}}"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for whitespace-only name")
	}
}

func TestCreateTemplateHandler_Validate_WhitespaceType(t *testing.T) {
	h := &CreateTemplateHandler{}
	req := &CreateTemplateRequest{Name: "Order", Type: "  ", Body: "Your order"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for whitespace-only type")
	}
}

func TestCreateTemplateHandler_Validate_WhitespaceBody(t *testing.T) {
	h := &CreateTemplateHandler{}
	req := &CreateTemplateRequest{Name: "Order", Type: "email", Body: "\t\n  "}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for whitespace-only body")
	}
}

func TestCreateTemplateHandler_Validate_AllEmpty(t *testing.T) {
	h := &CreateTemplateHandler{}
	req := &CreateTemplateRequest{}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when all fields are empty")
	}
}

func TestCreateTemplateHandler_Validate_WithSubjectPasses(t *testing.T) {
	h := &CreateTemplateHandler{}
	req := &CreateTemplateRequest{
		Name:    "Order Confirmation",
		Type:    "email",
		Body:    "Your order #{{id}} is confirmed.",
		Subject: "Order confirmed",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateTemplateHandler_Validate_RequiresID(t *testing.T) {
	h := &UpdateTemplateHandler{}
	req := &UpdateTemplateRequest{ID: 0, Name: "Updated"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when ID is zero")
	}
}

func TestUpdateTemplateHandler_Validate_RequiresName(t *testing.T) {
	h := &UpdateTemplateHandler{}
	req := &UpdateTemplateRequest{ID: 3, Name: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when name is empty")
	}
}

func TestUpdateTemplateHandler_Validate_HappyPath(t *testing.T) {
	h := &UpdateTemplateHandler{}
	req := &UpdateTemplateRequest{ID: 5, Name: "Updated Template"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeleteTemplateHandler_Validate_RequiresID(t *testing.T) {
	h := &DeleteTemplateHandler{}
	req := &DeleteTemplateRequest{ID: 0}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when ID is zero")
	}
}

func TestDeleteTemplateHandler_Validate_HappyPath(t *testing.T) {
	h := &DeleteTemplateHandler{}
	req := &DeleteTemplateRequest{ID: 7}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeleteTemplateHandler_Validate_LargeIDPasses(t *testing.T) {
	h := &DeleteTemplateHandler{}
	req := &DeleteTemplateRequest{ID: 99999}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error for large id: %v", err)
	}
}
