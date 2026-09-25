package domain

// OrderTotals holds computed financial totals for an order.
type OrderTotals struct {
	Subtotal       float64
	TaxAmount      float64
	DiscountAmount float64
	Total          float64
}

// ComputeTotals derives totals from a slice of order item inputs.
// Tax is applied after both percentage and flat per-item discounts.
func ComputeTotals(items []OrderItemInput) OrderTotals {
	var subtotal, taxAmount float64
	for _, item := range items {
		lineSubtotal := item.UnitPrice * item.Quantity
		lineDiscount := lineSubtotal * (item.DiscountPct / 100)
		lineAfterPercentage := lineSubtotal - lineDiscount
		flatDiscount := item.DiscountAmount
		if flatDiscount < 0 {
			flatDiscount = 0
		}
		if flatDiscount > lineAfterPercentage {
			flatDiscount = lineAfterPercentage
		}
		lineAfterDiscount := lineAfterPercentage - flatDiscount
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
