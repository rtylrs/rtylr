package domain

import (
	"fmt"
	"strings"
	"time"
)

// ValidateInvoiceInput returns a non-nil error when the input violates any
// business rule that can be checked without hitting the database.
func ValidateInvoiceInput(input InvoiceInput) error {
	if strings.TrimSpace(input.ClientName) == "" {
		return fmt.Errorf("client_name is required")
	}

	if strings.TrimSpace(input.IssueDate) == "" {
		return fmt.Errorf("issue_date is required")
	}

	if strings.TrimSpace(input.DueDate) == "" {
		return fmt.Errorf("due_date is required")
	}

	issue, err := time.Parse("2006-01-02", input.IssueDate)
	if err != nil {
		return fmt.Errorf("issue_date must be in YYYY-MM-DD format")
	}

	due, err := time.Parse("2006-01-02", input.DueDate)
	if err != nil {
		return fmt.Errorf("due_date must be in YYYY-MM-DD format")
	}

	if due.Before(issue) {
		return ErrInvalidInvoiceDate
	}

	if input.TaxRate < 0 || input.TaxRate > 100 {
		return fmt.Errorf("tax_rate must be between 0 and 100")
	}

	if input.DiscountValue < 0 {
		return fmt.Errorf("discount_value must not be negative")
	}

	if input.DiscountType != "" && input.DiscountType != "fixed" && input.DiscountType != "percentage" {
		return fmt.Errorf("discount_type must be 'fixed' or 'percentage'")
	}

	for i, item := range input.Items {
		if strings.TrimSpace(item.Description) == "" {
			return fmt.Errorf("item[%d]: description is required", i)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("item[%d]: quantity must be positive", i)
		}
		if item.UnitPrice < 0 {
			return fmt.Errorf("item[%d]: unit_price must not be negative", i)
		}
	}

	return nil
}

// ValidatePaymentInput returns a non-nil error when the payment input is
// structurally invalid.
func ValidatePaymentInput(input PaymentInput) error {
	if input.Amount <= 0 {
		return fmt.Errorf("payment amount must be positive")
	}

	if strings.TrimSpace(input.PaymentDate) == "" {
		return fmt.Errorf("payment_date is required")
	}

	if _, err := time.Parse("2006-01-02", input.PaymentDate); err != nil {
		return fmt.Errorf("payment_date must be in YYYY-MM-DD format")
	}

	if strings.TrimSpace(input.Method) == "" {
		return fmt.Errorf("payment method is required")
	}

	return nil
}
