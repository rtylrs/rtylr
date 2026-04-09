package domain

// PLReport holds Profit & Loss report data for a given period.
type PLReport struct {
	Period            string  `json:"period"`
	Revenue           float64 `json:"revenue"`
	COGS              float64 `json:"cogs"`
	GrossProfit       float64 `json:"grossProfit"`
	OperatingExpenses float64 `json:"operatingExpenses"`
	NetProfit         float64 `json:"netProfit"`
	MarginPct         float64 `json:"marginPct"`
}

// Dashboard holds the finance dashboard summary.
type Dashboard struct {
	TotalReceivable  float64 `json:"totalReceivable"`
	OverdueAmount    float64 `json:"overdueAmount"`
	TotalExpensesMTD float64 `json:"totalExpensesMtd"`
	NetProfitMTD     float64 `json:"netProfitMtd"`
}

// TaxReport holds tax data for a given period.
type TaxReport struct {
	Period       string  `json:"period"`
	TaxCollected float64 `json:"taxCollected"`
	TaxOwed      float64 `json:"taxOwed"`
}
