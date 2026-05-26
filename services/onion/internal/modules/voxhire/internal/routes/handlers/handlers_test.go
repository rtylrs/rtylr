package handlers

import (
	"context"
	"testing"
)

func TestContactHandler_Validate_RequiresCompanyName(t *testing.T) {
	h := &ContactHandler{}
	req := &ContactRequest{
		CompanyName: "",
		YourName:    "Alice",
		WorkEmail:   "alice@acme.com",
		Role:        "CTO",
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when company_name is empty")
	}
}

func TestContactHandler_Validate_RequiresYourName(t *testing.T) {
	h := &ContactHandler{}
	req := &ContactRequest{
		CompanyName: "Acme",
		YourName:    "",
		WorkEmail:   "alice@acme.com",
		Role:        "CTO",
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when your_name is empty")
	}
}

func TestContactHandler_Validate_RequiresWorkEmail(t *testing.T) {
	h := &ContactHandler{}
	req := &ContactRequest{
		CompanyName: "Acme",
		YourName:    "Alice",
		WorkEmail:   "",
		Role:        "CTO",
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when work_email is empty")
	}
}

func TestContactHandler_Validate_RequiresRole(t *testing.T) {
	h := &ContactHandler{}
	req := &ContactRequest{
		CompanyName: "Acme",
		YourName:    "Alice",
		WorkEmail:   "alice@acme.com",
		Role:        "",
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when role is empty")
	}
}

func TestContactHandler_Validate_HappyPath(t *testing.T) {
	h := &ContactHandler{}
	req := &ContactRequest{
		CompanyName: "Acme Corp",
		YourName:    "Alice Smith",
		WorkEmail:   "alice@acme.com",
		Role:        "Engineering Manager",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestContactHandler_Validate_AllEmpty(t *testing.T) {
	h := &ContactHandler{}
	req := &ContactRequest{}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when all fields are empty")
	}
}

func TestContactHandler_Validate_WithOptionalFields(t *testing.T) {
	h := &ContactHandler{}
	req := &ContactRequest{
		CompanyName: "StartupCo",
		YourName:    "Bob",
		WorkEmail:   "bob@startup.io",
		Role:        "CEO",
		Headcount:   "10-50",
		Timeline:    "Q3 2026",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error with optional fields: %v", err)
	}
}
