package domain

import (
	"testing"
	"time"
)

func TestValidateMenuItemInput_RejectsBlankName(t *testing.T) {
	if err := ValidateMenuItemInput(MenuItemInput{SalePrice: 5}); err == nil {
		t.Fatal("expected error for blank name")
	}
}

func TestValidateMenuItemInput_RejectsNegativePrice(t *testing.T) {
	if err := ValidateMenuItemInput(MenuItemInput{Name: "Burger", SalePrice: -1}); err == nil {
		t.Fatal("expected error for negative price")
	}
}

func TestValidateMenuItemInput_AcceptsZeroPrice(t *testing.T) {
	if err := ValidateMenuItemInput(MenuItemInput{Name: "Free Item", SalePrice: 0}); err != nil {
		t.Fatalf("unexpected error for zero price: %v", err)
	}
}

func TestValidateMenuItemInput_AcceptsValid(t *testing.T) {
	if err := ValidateMenuItemInput(MenuItemInput{Name: "Pizza", SalePrice: 12.99}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMenuItem_IsAvailable_Active(t *testing.T) {
	m := MenuItem{IsActive: true, SalePrice: 5, Name: "Item", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if !m.IsAvailable() {
		t.Error("expected IsAvailable=true for active item with valid price")
	}
}

func TestMenuItem_IsAvailable_Inactive(t *testing.T) {
	m := MenuItem{IsActive: false, SalePrice: 5}
	if m.IsAvailable() {
		t.Error("expected IsAvailable=false for inactive item")
	}
}

func TestMenuItem_IsAvailable_ZeroPrice(t *testing.T) {
	m := MenuItem{IsActive: true, SalePrice: 0}
	if !m.IsAvailable() {
		t.Error("expected IsAvailable=true for active item with zero price")
	}
}
