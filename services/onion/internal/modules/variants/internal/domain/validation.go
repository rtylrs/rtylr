package domain

import "errors"

// Validate returns an error if the VariantInput is invalid.
func (i VariantInput) Validate() error {
	if i.Name == "" {
		return errors.New("variant name is required")
	}
	if i.ProductID <= 0 {
		return errors.New("variant must be linked to a valid product")
	}
	if i.SalePrice < 0 {
		return errors.New("sale price must not be negative")
	}
	if i.CostPrice < 0 {
		return errors.New("cost price must not be negative")
	}
	if i.StockQuantity < 0 {
		return errors.New("stock quantity must not be negative")
	}
	return nil
}
