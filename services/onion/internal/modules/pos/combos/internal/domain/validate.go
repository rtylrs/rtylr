package domain

import "fmt"

// ValidateComboInput returns an error when the combo input is malformed.
func ValidateComboInput(input ComboInput) error {
	if input.Name == "" {
		return fmt.Errorf("combo name is required")
	}
	if input.Price < 0 {
		return fmt.Errorf("combo price cannot be negative")
	}
	return nil
}

// IsAvailable reports whether the combo is active and priced correctly.
func (c Combo) IsAvailable() bool {
	return c.IsActive && c.Price >= 0
}

// DiscountVsTotal computes the savings when comparing combo price against a full total.
// Returns 0 when the combo is not cheaper.
func (c Combo) DiscountVsTotal(fullPrice float64) float64 {
	if fullPrice <= c.Price {
		return 0
	}
	return fullPrice - c.Price
}
