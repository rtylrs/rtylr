package domain

type DaySummary struct {
	TotalSales       float64            `json:"total_sales"`
	TotalTax         float64            `json:"total_tax"`
	TotalTips        float64            `json:"total_tips"`
	TotalRefunds     float64            `json:"total_refunds"`
	TotalOrders      int                `json:"total_orders"`
	PaymentBreakdown map[string]float64 `json:"payment_breakdown"`
}

// CentsToFloat converts integer cents to standard float amount.
func CentsToFloat(cents int64) float64 {
	return float64(cents) / 100.0
}
