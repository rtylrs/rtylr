package api

import (
	"context"
	"testing"
)

func TestContactHandler_Validate_RequiresName(t *testing.T) {
	h := &ContactHandler{}
	req := &ContactRequest{Name: "", Email: "alice@example.com"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when name is empty")
	}
}

func TestContactHandler_Validate_RequiresEmail(t *testing.T) {
	h := &ContactHandler{}
	req := &ContactRequest{Name: "Alice", Email: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when email is empty")
	}
}

func TestContactHandler_Validate_HappyPath(t *testing.T) {
	h := &ContactHandler{}
	req := &ContactRequest{Name: "Alice", Email: "alice@example.com"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestContactHandler_Validate_BothEmpty(t *testing.T) {
	h := &ContactHandler{}
	req := &ContactRequest{}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when both fields are empty")
	}
}

func TestContactHandler_Validate_WithMessage(t *testing.T) {
	h := &ContactHandler{}
	req := &ContactRequest{
		Name:    "Bob",
		Email:   "bob@example.com",
		Message: "I need help with my project",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestContactHandler_Validate_WithPhone(t *testing.T) {
	h := &ContactHandler{}
	req := &ContactRequest{
		Name:        "Carol",
		Email:       "carol@example.com",
		PhoneNumber: "+9613940708",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSubmitHandler_Validate_RequiresName(t *testing.T) {
	h := &SubmitHandler{}
	req := &SubmitRequest{Name: "", Email: "alice@example.com", Service: "web"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when name is empty")
	}
}

func TestSubmitHandler_Validate_RequiresEmail(t *testing.T) {
	h := &SubmitHandler{}
	req := &SubmitRequest{Name: "Alice", Email: "", Service: "web"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when email is empty")
	}
}

func TestSubmitHandler_Validate_RequiresService(t *testing.T) {
	h := &SubmitHandler{}
	req := &SubmitRequest{Name: "Alice", Email: "alice@example.com", Service: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when service is empty")
	}
}

func TestSubmitHandler_Validate_HappyPath(t *testing.T) {
	h := &SubmitHandler{}
	req := &SubmitRequest{
		Name:    "Alice",
		Email:   "alice@example.com",
		Service: "web-design",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSubmitHandler_Validate_AllEmpty(t *testing.T) {
	h := &SubmitHandler{}
	req := &SubmitRequest{}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when all fields are empty")
	}
}

func TestSubmitHandler_Validate_WithBudget(t *testing.T) {
	h := &SubmitHandler{}
	req := &SubmitRequest{
		Name:    "Dave",
		Email:   "dave@example.com",
		Service: "branding",
		Budget:  "$5,000 - $10,000",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSubmitHandler_Validate_WithAllFields(t *testing.T) {
	h := &SubmitHandler{}
	req := &SubmitRequest{
		Name:     "Eve",
		Email:    "eve@example.com",
		Service:  "erp",
		Company:  "TechCorp",
		Project:  "Need full ERP setup",
		Budget:   "$10,000+",
		Timeline: "3 months",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error with all fields: %v", err)
	}
}
