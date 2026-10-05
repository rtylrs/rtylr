package domain

import "fmt"

// ValidateModifierInput returns an error when the input is malformed.
func ValidateModifierInput(input ModifierInput) error {
	if input.Name == "" {
		return fmt.Errorf("modifier name is required")
	}
	return nil
}

// ApplyTo returns the adjusted price after applying the modifier.
func (m Modifier) ApplyTo(basePrice float64) float64 {
	return basePrice + m.PriceDelta
}

// IsFree reports whether this modifier adds no cost.
func (m Modifier) IsFree() bool {
	return m.PriceDelta == 0
}

// IsUpcharge reports whether this modifier adds to the price.
func (m Modifier) IsUpcharge() bool {
	return m.PriceDelta > 0
}
