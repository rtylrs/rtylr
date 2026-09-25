package domain

import "testing"

func TestSupplierInput_Validate_EmptyName(t *testing.T) {
	err := (SupplierInput{Name: ""}).Validate()
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestSupplierInput_Validate_WhitespaceName(t *testing.T) {
	err := (SupplierInput{Name: "   "}).Validate()
	if err == nil {
		t.Fatal("expected error for whitespace name, got nil")
	}
}

func TestSupplierInput_Validate_ValidName(t *testing.T) {
	err := (SupplierInput{Name: "Acme Corp"}).Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSupplierInput_Validate_ValidEmail(t *testing.T) {
	err := (SupplierInput{Name: "Acme", Email: "acme@example.com"}).Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSupplierInput_Validate_InvalidEmail(t *testing.T) {
	err := (SupplierInput{Name: "Acme", Email: "not-an-email"}).Validate()
	if err == nil {
		t.Fatal("expected error for invalid email, got nil")
	}
}

func TestSupplierInput_Validate_EmptyEmailIsOK(t *testing.T) {
	err := (SupplierInput{Name: "Acme", Email: ""}).Validate()
	if err != nil {
		t.Fatalf("empty email should be allowed, got: %v", err)
	}
}
