package domain

import "testing"

func TestVariant_Status_Active(t *testing.T) {
	v := Variant{IsActive: true}
	if got := v.Status(); got != VariantStatusActive {
		t.Errorf("expected %q, got %q", VariantStatusActive, got)
	}
}

func TestVariant_Status_Inactive(t *testing.T) {
	v := Variant{IsActive: false}
	if got := v.Status(); got != VariantStatusInactive {
		t.Errorf("expected %q, got %q", VariantStatusInactive, got)
	}
}

func TestVariantStatusActive_StringValue(t *testing.T) {
	if string(VariantStatusActive) != "active" {
		t.Errorf("expected 'active', got %q", VariantStatusActive)
	}
}

func TestVariantStatusInactive_StringValue(t *testing.T) {
	if string(VariantStatusInactive) != "inactive" {
		t.Errorf("expected 'inactive', got %q", VariantStatusInactive)
	}
}

func TestVariantStatus_Distinct(t *testing.T) {
	if VariantStatusActive == VariantStatusInactive {
		t.Error("active and inactive must be distinct")
	}
}
