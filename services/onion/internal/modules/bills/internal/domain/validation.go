package domain

import (
	"fmt"
	"strings"
	"time"
)

// ValidBillStatuses is the set of statuses accepted by the bills module.
var ValidBillStatuses = map[string]struct{}{
	"pending": {},
	"paid":    {},
	"overdue": {},
}

// ValidateBillInput returns a non-nil error when the input violates any
// rule that can be checked without hitting the database.
func ValidateBillInput(input BillInput) error {
	if input.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	if strings.TrimSpace(input.DueDate) == "" {
		return fmt.Errorf("due_date is required")
	}

	if _, err := time.Parse("2006-01-02", input.DueDate); err != nil {
		return fmt.Errorf("due_date must be in YYYY-MM-DD format")
	}

	if input.Status != "" {
		if _, ok := ValidBillStatuses[input.Status]; !ok {
			return fmt.Errorf("invalid status %q: must be one of pending, paid, overdue", input.Status)
		}
	}

	return nil
}

// ValidateBillCreateInput requires every new bill to enter the workflow as
// pending. Overdue is assigned by aging workflows, and paid must follow a
// payment transition.
func ValidateBillCreateInput(input BillInput) error {
	if input.Status != "" && input.Status != "pending" {
		return fmt.Errorf("new bills must start with pending status")
	}
	return ValidateBillInput(input)
}
