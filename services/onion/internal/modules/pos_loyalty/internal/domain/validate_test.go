package domain

import "testing"

func TestValidateEarnPointsInput_RejectsMissingBusinessID(t *testing.T) {
	if err := ValidateEarnPointsInput(EarnPointsInput{CustomerID: 1, Points: 10}); err == nil {
		t.Fatal("expected error for missing business id")
	}
}

func TestValidateEarnPointsInput_RejectsMissingCustomerID(t *testing.T) {
	if err := ValidateEarnPointsInput(EarnPointsInput{BusinessID: 1, Points: 10}); err == nil {
		t.Fatal("expected error for missing customer id")
	}
}

func TestValidateEarnPointsInput_RejectsZeroPoints(t *testing.T) {
	if err := ValidateEarnPointsInput(EarnPointsInput{BusinessID: 1, CustomerID: 2, Points: 0}); err == nil {
		t.Fatal("expected error for zero points")
	}
}

func TestValidateEarnPointsInput_RejectsNegativePoints(t *testing.T) {
	if err := ValidateEarnPointsInput(EarnPointsInput{BusinessID: 1, CustomerID: 2, Points: -5}); err == nil {
		t.Fatal("expected error for negative points")
	}
}

func TestValidateEarnPointsInput_AcceptsValid(t *testing.T) {
	if err := ValidateEarnPointsInput(EarnPointsInput{BusinessID: 1, CustomerID: 2, Points: 10}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRedeemPointsInput_RejectsMissingBusinessID(t *testing.T) {
	if err := ValidateRedeemPointsInput(RedeemPointsInput{CustomerID: 1, Points: 10}); err == nil {
		t.Fatal("expected error for missing business id")
	}
}

func TestValidateRedeemPointsInput_RejectsMissingCustomerID(t *testing.T) {
	if err := ValidateRedeemPointsInput(RedeemPointsInput{BusinessID: 1, Points: 10}); err == nil {
		t.Fatal("expected error for missing customer id")
	}
}

func TestValidateRedeemPointsInput_RejectsZeroPoints(t *testing.T) {
	if err := ValidateRedeemPointsInput(RedeemPointsInput{BusinessID: 1, CustomerID: 2, Points: 0}); err == nil {
		t.Fatal("expected error for zero points")
	}
}

func TestValidateRedeemPointsInput_AcceptsValid(t *testing.T) {
	if err := ValidateRedeemPointsInput(RedeemPointsInput{BusinessID: 1, CustomerID: 2, Points: 50}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
