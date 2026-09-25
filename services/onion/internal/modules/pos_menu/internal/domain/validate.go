package domain

import "fmt"

// ValidateMenuItemInput returns an error when the input is malformed.
func ValidateMenuItemInput(input MenuItemInput) error {
	if input.Name == "" {
		return fmt.Errorf("menu item name is required")
	}
	if input.SalePrice < 0 {
		return fmt.Errorf("sale price cannot be negative")
	}
	return nil
}

// IsAvailable reports whether the item can be sold.
func (m MenuItem) IsAvailable() bool {
	return m.IsActive && m.SalePrice >= 0
}
