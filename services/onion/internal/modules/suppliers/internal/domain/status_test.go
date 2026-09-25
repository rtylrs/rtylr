package domain

import "testing"

func TestSupplier_Status_Active(t *testing.T) {
	s := Supplier{IsActive: true}
	if got := s.Status(); got != SupplierStatusActive {
		t.Errorf("expected %q, got %q", SupplierStatusActive, got)
	}
}

func TestSupplier_Status_Inactive(t *testing.T) {
	s := Supplier{IsActive: false}
	if got := s.Status(); got != SupplierStatusInactive {
		t.Errorf("expected %q, got %q", SupplierStatusInactive, got)
	}
}

func TestSupplierStatusActive_StringValue(t *testing.T) {
	if string(SupplierStatusActive) != "active" {
		t.Errorf("expected 'active', got %q", SupplierStatusActive)
	}
}

func TestSupplierStatusInactive_StringValue(t *testing.T) {
	if string(SupplierStatusInactive) != "inactive" {
		t.Errorf("expected 'inactive', got %q", SupplierStatusInactive)
	}
}

func TestSupplierStatus_Distinct(t *testing.T) {
	if SupplierStatusActive == SupplierStatusInactive {
		t.Error("active and inactive statuses must be distinct")
	}
}
