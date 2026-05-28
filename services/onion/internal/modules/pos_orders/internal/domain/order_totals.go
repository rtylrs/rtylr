package domain

// OrderTotals holds computed financial totals for an order.
type OrderTotals struct {
	Subtotal       float64
	TaxAmount      float64
	DiscountAmount float64
	Total          float64
}

// ComputeTotals derives totals from a slice of order item inputs.
// Tax is applied after the per-item flat discount (not percentage discounts,
// which are resolved by the application layer before calling this).
func ComputeTotals(items []OrderItemInput) OrderTotals {
	var subtotal, taxAmount float64
	for _, item := range items {
		lineSubtotal := item.UnitPrice * item.Quantity
		lineDiscount := lineSubtotal * (item.DiscountPct / 100)
		lineAfterDiscount := lineSubtotal - lineDiscount
		lineTax := lineAfterDiscount * (item.TaxRate / 100)
		subtotal += lineAfterDiscount
		taxAmount += lineTax
	}
	return OrderTotals{
		Subtotal:  subtotal,
		TaxAmount: taxAmount,
		Total:     subtotal + taxAmount,
	}
}
