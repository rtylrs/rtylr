package domain

type KPISummary struct {
	Revenue       float64 `json:"revenue"`
	OrderCount    int64   `json:"orderCount"`
	StaffCount    int64   `json:"staffCount"`
	CustomerCount int64   `json:"customerCount"`
	LowStockCount int64   `json:"lowStockCount"`
	ActiveApps    int64   `json:"activeApps"`
}
