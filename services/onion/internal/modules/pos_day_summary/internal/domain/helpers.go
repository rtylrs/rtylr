package domain

// NetRevenue computes net revenue after refunds.
func (d DaySummary) NetRevenue() float64 {
	return d.TotalSales - d.TotalRefunds
}

// NetSalesAfterTax returns the sales amount excluding tax.
func (d DaySummary) NetSalesAfterTax() float64 {
	return d.TotalSales - d.TotalTax
}

// AverageOrderValue returns the average order value, or 0 when there are no orders.
func (d DaySummary) AverageOrderValue() float64 {
	if d.TotalOrders == 0 {
		return 0
	}
	return d.TotalSales / float64(d.TotalOrders)
}

// HasRefunds reports whether any refunds were issued during the day.
func (d DaySummary) HasRefunds() bool {
	return d.TotalRefunds > 0
}

// CashSales returns the cash payment amount from the payment breakdown.
func (d DaySummary) CashSales() float64 {
	if d.PaymentBreakdown == nil {
		return 0
	}
	return d.PaymentBreakdown["cash"]
}

// CardSales returns the card payment amount from the payment breakdown.
func (d DaySummary) CardSales() float64 {
	if d.PaymentBreakdown == nil {
		return 0
	}
	return d.PaymentBreakdown["card"]
}
