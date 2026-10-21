package domain

import "testing"

func TestSupplier_HasContactInfo_EmailOnly(t *testing.T) {
	s := Supplier{Name: "Acme", Email: "info@acme.com"}
	if !s.HasContactInfo() {
		t.Error("expected true when email is set")
	}
}

func TestSupplier_HasContactInfo_PhoneOnly(t *testing.T) {
	s := Supplier{Name: "Acme", Phone: "+1234567890"}
	if !s.HasContactInfo() {
		t.Error("expected true when phone is set")
	}
}

func TestSupplier_HasContactInfo_BothSet(t *testing.T) {
	s := Supplier{Name: "Acme", Email: "a@b.com", Phone: "123"}
	if !s.HasContactInfo() {
		t.Error("expected true when both email and phone are set")
	}
}

func TestSupplier_HasContactInfo_NeitherSet(t *testing.T) {
	s := Supplier{Name: "Acme"}
	if s.HasContactInfo() {
		t.Error("expected false when neither email nor phone is set")
	}
}

func TestSupplier_HasAddress_BothSet(t *testing.T) {
	s := Supplier{Address: "123 Main St", City: "Testville"}
	if !s.HasAddress() {
		t.Error("expected true when address and city are set")
	}
}

func TestSupplier_HasAddress_MissingCity(t *testing.T) {
	s := Supplier{Address: "123 Main St", City: ""}
	if s.HasAddress() {
		t.Error("expected false when city is missing")
	}
}

func TestSupplier_HasAddress_MissingAddress(t *testing.T) {
	s := Supplier{Address: "", City: "Testville"}
	if s.HasAddress() {
		t.Error("expected false when address is missing")
	}
}

func TestSupplier_HasAddress_BothEmpty(t *testing.T) {
	s := Supplier{}
	if s.HasAddress() {
		t.Error("expected false when both address and city are empty")
	}
}

func TestSupplier_DisplayName_Trimmed(t *testing.T) {
	s := Supplier{Name: "  Acme Corp  "}
	if got := s.DisplayName(); got != "Acme Corp" {
		t.Errorf("expected 'Acme Corp', got %q", got)
	}
}

func TestSupplier_DisplayName_NoTrim(t *testing.T) {
	s := Supplier{Name: "Clean Name"}
	if got := s.DisplayName(); got != "Clean Name" {
		t.Errorf("expected 'Clean Name', got %q", got)
	}
}
