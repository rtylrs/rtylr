package domain

import (
	"fmt"
	"strings"
	"time"
)

// ValidQuoteStatuses lists the statuses accepted by the quotes module.
var ValidQuoteStatuses = map[string]struct{}{
	"draft":    {},
	"sent":     {},
	"accepted": {},
	"declined": {},
	"expired":  {},
}

// ValidateQuoteInput returns a non-nil error when the input violates a
// business rule that can be checked without hitting the database.
func ValidateQuoteInput(input QuoteInput) error {
	if strings.TrimSpace(input.CustomerName) == "" {
		return fmt.Errorf("customer_name is required")
	}

	if input.Total < 0 {
		return fmt.Errorf("total must not be negative")
	}

	if input.Subtotal < 0 {
		return fmt.Errorf("subtotal must not be negative")
	}

	if input.Tax < 0 {
		return fmt.Errorf("tax must not be negative")
	}

	if input.Discount < 0 {
		return fmt.Errorf("discount must not be negative")
	}

	if input.Status != "" {
		if _, ok := ValidQuoteStatuses[input.Status]; !ok {
			return fmt.Errorf("invalid status %q", input.Status)
		}
	}

	if input.ValidUntil != "" {
		if _, err := time.Parse("2006-01-02", input.ValidUntil); err != nil {
			return fmt.Errorf("valid_until must be in YYYY-MM-DD format")
		}
	}

	return nil
}
