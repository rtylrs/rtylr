package domain

import "testing"

func TestValidateModifierInput_RejectsBlankName(t *testing.T) {
	if err := ValidateModifierInput(ModifierInput{}); err == nil {
		t.Fatal("expected error for blank name")
	}
}

func TestValidateModifierInput_AcceptsValid(t *testing.T) {
	if err := ValidateModifierInput(ModifierInput{Name: "Extra Cheese"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestModifier_ApplyTo_Upcharge(t *testing.T) {
	m := Modifier{PriceDelta: 2.50}
	got := m.ApplyTo(10.00)
	if got != 12.50 {
		t.Errorf("ApplyTo: got %g, want 12.50", got)
	}
}

func TestModifier_ApplyTo_Discount(t *testing.T) {
	m := Modifier{PriceDelta: -1.00}
	got := m.ApplyTo(10.00)
	if got != 9.00 {
		t.Errorf("ApplyTo: got %g, want 9.00", got)
	}
}

func TestModifier_ApplyTo_Free(t *testing.T) {
	m := Modifier{PriceDelta: 0}
	got := m.ApplyTo(5.00)
	if got != 5.00 {
		t.Errorf("ApplyTo: got %g, want 5.00", got)
	}
}

func TestModifier_IsFree_ZeroDelta(t *testing.T) {
	m := Modifier{PriceDelta: 0}
	if !m.IsFree() {
		t.Error("expected IsFree=true for zero delta")
	}
}

func TestModifier_IsFree_NonZeroDelta(t *testing.T) {
	m := Modifier{PriceDelta: 1.00}
	if m.IsFree() {
		t.Error("expected IsFree=false for non-zero delta")
	}
}

func TestModifier_IsUpcharge_PositiveDelta(t *testing.T) {
	m := Modifier{PriceDelta: 0.50}
	if !m.IsUpcharge() {
		t.Error("expected IsUpcharge=true for positive delta")
	}
}

func TestModifier_IsUpcharge_NegativeDelta(t *testing.T) {
	m := Modifier{PriceDelta: -1.00}
	if m.IsUpcharge() {
		t.Error("expected IsUpcharge=false for negative delta")
	}
}

func TestModifier_IsUpcharge_ZeroDelta(t *testing.T) {
	m := Modifier{PriceDelta: 0}
	if m.IsUpcharge() {
		t.Error("expected IsUpcharge=false for zero delta")
	}
}
