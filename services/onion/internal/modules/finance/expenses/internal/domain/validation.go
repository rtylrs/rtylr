package domain

import (
	"fmt"
	"strings"
	"time"
)

// ValidExpenseStatuses is the set of statuses accepted by the expenses module.
var ValidExpenseStatuses = map[string]struct{}{
	"pending":  {},
	"approved": {},
	"rejected": {},
}

// ValidPaymentMethods mirrors the finance_expenses.payment_method enum. An
// empty value is stored as the column default, cash.
var ValidPaymentMethods = map[string]struct{}{
	"cash":          {},
	"card":          {},
	"bank_transfer": {},
	"other":         {},
}

// ValidatePaymentMethod refuses a value the column would reject, so a bad
// request is a 400 rather than a failed insert.
func ValidatePaymentMethod(method string) error {
	if method == "" {
		return nil
	}
	if _, ok := ValidPaymentMethods[method]; !ok {
		return fmt.Errorf("payment_method must be one of cash, card, bank_transfer or other")
	}
	return nil
}

// ValidatExpenseInput returns a non-nil error when the input fails basic
// structural checks that can be evaluated without a database round-trip.
func ValidatExpenseInput(amount float64, currency, description string, expenseDate time.Time) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	if expenseDate.IsZero() {
		return fmt.Errorf("expense_date is required")
	}

	if expenseDate.After(time.Now().AddDate(0, 0, 1)) {
		return fmt.Errorf("expense_date must not be in the future")
	}

	if strings.TrimSpace(description) == "" {
		return fmt.Errorf("description is required")
	}

	if currency != "" && len(currency) != 3 {
		return fmt.Errorf("currency must be a 3-letter ISO code")
	}

	return nil
}

// ValidateCategoryInput checks that a category name is non-empty.
func ValidateCategoryInput(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("category name is required")
	}

	return nil
}
