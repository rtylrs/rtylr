package domain

type KPISummary struct {
	Revenue       float64 `json:"revenue"`
	OrderCount    int64   `json:"order_count"`
	StaffCount    int64   `json:"staff_count"`
	CustomerCount int64   `json:"customer_count"`
	LowStockCount int64   `json:"low_stock_count"`
	ActiveApps    int64   `json:"active_apps"`
}
