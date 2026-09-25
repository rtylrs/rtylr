package domain

import "errors"

// Validate returns an error if the ProductInput is invalid.
func (i ProductInput) Validate() error {
	if i.Name == "" {
		return errors.New("product name is required")
	}
	if i.SalePrice < 0 {
		return errors.New("sale price must not be negative")
	}
	if i.CostPrice < 0 {
		return errors.New("cost price must not be negative")
	}
	if i.LowStockThreshold < 0 {
		return errors.New("low stock threshold must not be negative")
	}
	if i.ReorderQuantity < 0 {
		return errors.New("reorder quantity must not be negative")
	}
	return nil
}
