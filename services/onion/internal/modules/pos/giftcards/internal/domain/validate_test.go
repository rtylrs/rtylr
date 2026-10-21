package domain

import "testing"

func TestValidateGiftCardInput_RejectsBlankCode(t *testing.T) {
	if err := ValidateGiftCardInput(GiftCardInput{InitialAmount: 50}); err == nil {
		t.Fatal("expected error for blank code")
	}
}

func TestValidateGiftCardInput_RejectsZeroAmount(t *testing.T) {
	if err := ValidateGiftCardInput(GiftCardInput{Code: "GC-001", InitialAmount: 0}); err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestValidateGiftCardInput_RejectsNegativeAmount(t *testing.T) {
	if err := ValidateGiftCardInput(GiftCardInput{Code: "GC-001", InitialAmount: -10}); err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestValidateGiftCardInput_AcceptsValid(t *testing.T) {
	if err := ValidateGiftCardInput(GiftCardInput{PaidOrderID: 1, Code: "GC-001", InitialAmount: 100}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRedeemInput_RejectsBlankCode(t *testing.T) {
	if err := ValidateRedeemInput(RedeemInput{Amount: 10}); err == nil {
		t.Fatal("expected error for blank code")
	}
}

func TestValidateRedeemInput_RejectsZeroAmount(t *testing.T) {
	if err := ValidateRedeemInput(RedeemInput{Code: "GC-001", Amount: 0}); err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestValidateRedeemInput_RejectsNegativeAmount(t *testing.T) {
	if err := ValidateRedeemInput(RedeemInput{Code: "GC-001", Amount: -5}); err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestValidateRedeemInput_AcceptsValid(t *testing.T) {
	if err := ValidateRedeemInput(RedeemInput{OrderID: 1, Code: "GC-001", Amount: 25}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
