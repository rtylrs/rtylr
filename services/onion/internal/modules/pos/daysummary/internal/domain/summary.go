package domain

type DaySummary struct {
	TotalSales       float64
	TotalTax         float64
	TotalTips        float64
	TotalRefunds     float64
	TotalOrders      int
	PaymentBreakdown map[string]float64
}
