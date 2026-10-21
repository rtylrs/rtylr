package domain

import "testing"

func TestValidateComboInput_RejectsBlankName(t *testing.T) {
	if err := ValidateComboInput(ComboInput{Price: 10}); err == nil {
		t.Fatal("expected error for blank name")
	}
}

func TestValidateComboInput_RejectsNegativePrice(t *testing.T) {
	if err := ValidateComboInput(ComboInput{Name: "Meal Deal", Price: -5}); err == nil {
		t.Fatal("expected error for negative price")
	}
}

func TestValidateComboInput_AcceptsZeroPrice(t *testing.T) {
	if err := ValidateComboInput(ComboInput{Name: "Free Combo", Price: 0}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateComboInput_AcceptsValid(t *testing.T) {
	if err := ValidateComboInput(ComboInput{Name: "Burger Meal", Price: 14.99}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCombo_IsAvailable_Active(t *testing.T) {
	c := Combo{IsActive: true, Price: 10}
	if !c.IsAvailable() {
		t.Error("expected IsAvailable=true for active combo with valid price")
	}
}

func TestCombo_IsAvailable_Inactive(t *testing.T) {
	c := Combo{IsActive: false, Price: 10}
	if c.IsAvailable() {
		t.Error("expected IsAvailable=false for inactive combo")
	}
}

func TestCombo_DiscountVsTotal_Savings(t *testing.T) {
	c := Combo{Price: 12}
	if got := c.DiscountVsTotal(15); got != 3 {
		t.Errorf("DiscountVsTotal: got %g, want 3", got)
	}
}

func TestCombo_DiscountVsTotal_NoCheaper(t *testing.T) {
	c := Combo{Price: 20}
	if got := c.DiscountVsTotal(10); got != 0 {
		t.Errorf("DiscountVsTotal when combo costs more: got %g, want 0", got)
	}
}

func TestCombo_DiscountVsTotal_EqualPrice(t *testing.T) {
	c := Combo{Price: 10}
	if got := c.DiscountVsTotal(10); got != 0 {
		t.Errorf("DiscountVsTotal at same price: got %g, want 0", got)
	}
}
