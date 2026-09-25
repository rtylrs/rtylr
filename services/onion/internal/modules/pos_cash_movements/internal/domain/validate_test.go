package domain

import "testing"

func TestValidateCreateCashMovementInput_RejectsMissingBusinessID(t *testing.T) {
	if err := ValidateCreateCashMovementInput(CreateCashMovementInput{
		BranchID: 1, SessionID: 2, EmployeeID: 3, Type: "cash_in", Amount: 100,
	}); err == nil {
		t.Fatal("expected error for missing business id")
	}
}

func TestValidateCreateCashMovementInput_RejectsMissingBranchID(t *testing.T) {
	if err := ValidateCreateCashMovementInput(CreateCashMovementInput{
		BusinessID: 1, SessionID: 2, EmployeeID: 3, Type: "cash_in", Amount: 100,
	}); err == nil {
		t.Fatal("expected error for missing branch id")
	}
}

func TestValidateCreateCashMovementInput_RejectsMissingSessionID(t *testing.T) {
	if err := ValidateCreateCashMovementInput(CreateCashMovementInput{
		BusinessID: 1, BranchID: 2, EmployeeID: 3, Type: "cash_in", Amount: 100,
	}); err == nil {
		t.Fatal("expected error for missing session id")
	}
}

func TestValidateCreateCashMovementInput_RejectsMissingEmployeeID(t *testing.T) {
	if err := ValidateCreateCashMovementInput(CreateCashMovementInput{
		BusinessID: 1, BranchID: 2, SessionID: 3, Type: "cash_in", Amount: 100,
	}); err == nil {
		t.Fatal("expected error for missing employee id")
	}
}

func TestValidateCreateCashMovementInput_RejectsZeroAmount(t *testing.T) {
	if err := ValidateCreateCashMovementInput(CreateCashMovementInput{
		BusinessID: 1, BranchID: 2, SessionID: 3, EmployeeID: 4, Type: "cash_in", Amount: 0,
	}); err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestValidateCreateCashMovementInput_RejectsInvalidType(t *testing.T) {
	if err := ValidateCreateCashMovementInput(CreateCashMovementInput{
		BusinessID: 1, BranchID: 2, SessionID: 3, EmployeeID: 4, Type: "transfer", Amount: 50,
	}); err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestValidateCreateCashMovementInput_AcceptsCashIn(t *testing.T) {
	if err := ValidateCreateCashMovementInput(CreateCashMovementInput{
		BusinessID: 1, BranchID: 2, SessionID: 3, EmployeeID: 4, Type: "cash_in", Amount: 100,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateCreateCashMovementInput_AcceptsCashOut(t *testing.T) {
	if err := ValidateCreateCashMovementInput(CreateCashMovementInput{
		BusinessID: 1, BranchID: 2, SessionID: 3, EmployeeID: 4, Type: "cash_out", Amount: 50,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
