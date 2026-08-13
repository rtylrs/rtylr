package domain

// PLReport holds Profit & Loss report data for a given period.
type PLReport struct {
	Period            string  `json:"period"`
	Revenue           float64 `json:"revenue"`
	COGS              float64 `json:"cogs"`
	GrossProfit       float64 `json:"gross_profit"`
	OperatingExpenses float64 `json:"operating_expenses"`
	NetProfit         float64 `json:"net_profit"`
	MarginPct         float64 `json:"margin_pct"`
}

// Dashboard holds the finance dashboard summary.
type Dashboard struct {
	TotalReceivable  float64 `json:"total_receivable"`
	OverdueAmount    float64 `json:"overdue_amount"`
	TotalExpensesMTD float64 `json:"total_expenses_mtd"`
	TotalRevenueMTD  float64 `json:"total_revenue_mtd"`
	NetProfitMTD     float64 `json:"net_profit_mtd"`
}

// TaxReport holds tax data for a given period.
type TaxReport struct {
	Period       string  `json:"period"`
	TaxCollected float64 `json:"tax_collected"`
	TaxOwed      float64 `json:"tax_owed"`
}
