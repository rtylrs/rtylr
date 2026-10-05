package domain

// LineTotal returns the effective total cost for the item (quantity * unit cost).
// This may differ from the stored TotalCost if the record is stale.
func (i PurchaseOrderItem) LineTotal() float64 {
	return i.QuantityOrdered * i.UnitCost
}

// RemainingQuantity returns the quantity yet to be received.
func (i PurchaseOrderItem) RemainingQuantity() float64 {
	remaining := i.QuantityOrdered - i.QuantityReceived
	if remaining < 0 {
		return 0
	}
	return remaining
}

// IsFullyReceived reports whether the ordered quantity has been fully received.
func (i PurchaseOrderItem) IsFullyReceived() bool {
	return i.QuantityReceived >= i.QuantityOrdered
}
