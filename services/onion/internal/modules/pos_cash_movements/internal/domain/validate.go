package domain

import "fmt"

// ValidateCreateCashMovementInput returns an error when the input is malformed.
func ValidateCreateCashMovementInput(input CreateCashMovementInput) error {
	if input.BusinessID == 0 {
		return fmt.Errorf("business id is required")
	}
	if input.BranchID == 0 {
		return fmt.Errorf("branch id is required")
	}
	if input.SessionID == 0 {
		return fmt.Errorf("session id is required")
	}
	if input.EmployeeID == 0 {
		return fmt.Errorf("employee id is required")
	}
	if input.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if input.Type != "cash_in" && input.Type != "cash_out" {
		return fmt.Errorf("type must be 'cash_in' or 'cash_out', got %q", input.Type)
	}
	return nil
}
