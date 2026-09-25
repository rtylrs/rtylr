package domain

import "errors"

// Validate returns an error if the InventoryAdjustInput is invalid.
func (i InventoryAdjustInput) Validate() error {
	if i.ProductID <= 0 {
		return errors.New("product ID must be positive")
	}
	if i.Change == 0 {
		return errors.New("inventory change must be non-zero")
	}
	switch i.Type {
	case InventoryTypePurchase, InventoryTypeSale, InventoryTypeAdjustment,
		InventoryTypeReturn, InventoryTypeWaste:
		// valid
	default:
		return errors.New("unknown inventory transaction type")
	}
	return nil
}
