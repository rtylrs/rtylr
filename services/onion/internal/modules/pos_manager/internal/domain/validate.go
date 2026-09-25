package domain

import (
	"fmt"
	"strings"
)

// ValidatePIN returns an error when the PIN does not meet requirements.
// PINs must be 4-8 digits.
func ValidatePIN(pin string) error {
	trimmed := strings.TrimSpace(pin)
	if len(trimmed) < 4 {
		return fmt.Errorf("PIN must be at least 4 digits")
	}
	if len(trimmed) > 8 {
		return fmt.Errorf("PIN must be at most 8 digits")
	}
	for _, ch := range trimmed {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("PIN must contain only digits")
		}
	}
	return nil
}
