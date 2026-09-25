package domain

import "testing"

func TestValidateBillInput_HappyPath(t *testing.T) {
	input := BillInput{Amount: 100, DueDate: "2026-12-31", Status: "pending"}
	if err := ValidateBillInput(input); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateBillInput_ZeroAmount(t *testing.T) {
	input := BillInput{Amount: 0, DueDate: "2026-12-31"}
	if err := ValidateBillInput(input); err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestValidateBillInput_NegativeAmount(t *testing.T) {
	input := BillInput{Amount: -50, DueDate: "2026-12-31"}
	if err := ValidateBillInput(input); err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestValidateBillInput_MissingDueDate(t *testing.T) {
	input := BillInput{Amount: 100}
	if err := ValidateBillInput(input); err == nil {
		t.Fatal("expected error for missing due_date")
	}
}

func TestValidateBillInput_BadDueDateFormat(t *testing.T) {
	input := BillInput{Amount: 100, DueDate: "31-12-2026"}
	if err := ValidateBillInput(input); err == nil {
		t.Fatal("expected error for bad date format")
	}
}

func TestValidateBillInput_InvalidStatus(t *testing.T) {
	input := BillInput{Amount: 100, DueDate: "2026-12-31", Status: "unknown"}
	if err := ValidateBillInput(input); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestValidateBillInput_StatusPaid(t *testing.T) {
	input := BillInput{Amount: 100, DueDate: "2026-12-31", Status: "paid"}
	if err := ValidateBillInput(input); err != nil {
		t.Errorf("unexpected error for status=paid: %v", err)
	}
}

func TestValidateBillInput_StatusOverdue(t *testing.T) {
	input := BillInput{Amount: 100, DueDate: "2026-12-31", Status: "overdue"}
	if err := ValidateBillInput(input); err != nil {
		t.Errorf("unexpected error for status=overdue: %v", err)
	}
}

func TestValidateBillInput_EmptyStatusIsAccepted(t *testing.T) {
	input := BillInput{Amount: 100, DueDate: "2026-12-31"}
	if err := ValidateBillInput(input); err != nil {
		t.Errorf("unexpected error for empty status: %v", err)
	}
}

func TestValidateBillCreateInput_RejectsPaidStatus(t *testing.T) {
	input := BillInput{Amount: 100, DueDate: "2026-12-31", Status: "paid"}
	if err := ValidateBillCreateInput(input); err == nil {
		t.Fatal("expected paid bill creation to be rejected")
	}
}
