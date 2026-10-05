package domain

// PLReport holds Profit & Loss report data for a given period.
type PLReport struct {
	Period            string
	Revenue           float64
	COGS              float64
	GrossProfit       float64
	OperatingExpenses float64
	NetProfit         float64
	MarginPct         float64
}

// Dashboard holds the finance dashboard summary.
type Dashboard struct {
	TotalReceivable  float64
	OverdueAmount    float64
	TotalExpensesMTD float64
	TotalRevenueMTD  float64
	NetProfitMTD     float64
}

// TaxReport holds tax data for a given period.
type TaxReport struct {
	Period       string
	TaxCollected float64
	TaxOwed      float64
}
