package domain

// ComputedSubtotal returns the sum of all item totals.
func (po PurchaseOrder) ComputedSubtotal() float64 {
	var total float64
	for _, item := range po.Items {
		total += item.TotalCost
	}
	return total
}

// ComputedTotal returns subtotal plus tax.
func (po PurchaseOrder) ComputedTotal() float64 {
	return po.ComputedSubtotal() + po.TaxAmount
}

// IsCancellable reports whether the purchase order can be cancelled.
// Only draft and sent orders may be cancelled.
func (po PurchaseOrder) IsCancellable() bool {
	return po.Status == POStatusDraft || po.Status == POStatusSent
}

// IsEditable reports whether the purchase order can have its items modified.
// Only draft orders are editable.
func (po PurchaseOrder) IsEditable() bool {
	return po.Status == POStatusDraft
}

// ItemCount returns the number of line items on the order.
func (po PurchaseOrder) ItemCount() int {
	return len(po.Items)
}

// TotalQuantityOrdered returns the sum of all ordered quantities.
func (po PurchaseOrder) TotalQuantityOrdered() float64 {
	var qty float64
	for _, item := range po.Items {
		qty += item.QuantityOrdered
	}
	return qty
}
