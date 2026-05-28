package domain

import "errors"

// Validate returns an error if the PurchaseOrderInput is invalid.
func (i PurchaseOrderInput) Validate() error {
	if i.SupplierID <= 0 {
		return errors.New("purchase order must reference a valid supplier")
	}
	if i.PONumber == "" {
		return errors.New("purchase order number is required")
	}
	if len(i.Items) == 0 {
		return errors.New("purchase order must have at least one item")
	}
	if i.TaxAmount < 0 {
		return errors.New("tax amount must not be negative")
	}
	for idx, item := range i.Items {
		if item.ProductID <= 0 {
			return errors.New("each item must reference a valid product")
		}
		if item.QuantityOrdered <= 0 {
			return errors.New("item quantity must be positive")
		}
		if item.UnitCost < 0 {
			return errors.New("item unit cost must not be negative")
		}
		_ = idx
	}
	return nil
}

// Validate returns an error if the PurchaseOrderItemInput is invalid.
func (i PurchaseOrderItemInput) Validate() error {
	if i.ProductID <= 0 {
		return errors.New("item must reference a valid product")
	}
	if i.QuantityOrdered <= 0 {
		return errors.New("quantity ordered must be positive")
	}
	if i.UnitCost < 0 {
		return errors.New("unit cost must not be negative")
	}
	return nil
}
