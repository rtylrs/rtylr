package domain

import "testing"

func TestProduct_Status_Active(t *testing.T) {
	p := Product{IsActive: true}
	if got := p.Status(); got != ProductStatusActive {
		t.Errorf("expected %q, got %q", ProductStatusActive, got)
	}
}

func TestProduct_Status_Inactive(t *testing.T) {
	p := Product{IsActive: false}
	if got := p.Status(); got != ProductStatusInactive {
		t.Errorf("expected %q, got %q", ProductStatusInactive, got)
	}
}

func TestProductStatusActive_String(t *testing.T) {
	if string(ProductStatusActive) != "active" {
		t.Errorf("expected 'active', got %q", ProductStatusActive)
	}
}

func TestProductStatusInactive_String(t *testing.T) {
	if string(ProductStatusInactive) != "inactive" {
		t.Errorf("expected 'inactive', got %q", ProductStatusInactive)
	}
}
