package domain

type DaySummary struct {
	TotalSales       float64            `json:"totalSales"`
	TotalTax         float64            `json:"totalTax"`
	TotalTips        float64            `json:"totalTips"`
	TotalRefunds     float64            `json:"totalRefunds"`
	TotalOrders      int                `json:"totalOrders"`
	PaymentBreakdown map[string]float64 `json:"paymentBreakdown"`
}
