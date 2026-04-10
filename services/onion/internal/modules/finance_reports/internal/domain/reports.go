package domain

type BalanceSheet struct {
	TotalAssets      float64 `json:"totalAssets"`
	TotalLiabilities float64 `json:"totalLiabilities"`
	TotalEquity      float64 `json:"totalEquity"`
	CashBalance      float64 `json:"cashBalance"`
	AccountsReceivable float64 `json:"accountsReceivable"`
	AccountsPayable    float64 `json:"accountsPayable"`
}

type CashFlow struct {
	Period             string  `json:"period"`
	OperatingInflow    float64 `json:"operatingInflow"`
	OperatingOutflow   float64 `json:"operatingOutflow"`
	NetOperating       float64 `json:"netOperating"`
	NetCashFlow        float64 `json:"netCashFlow"`
}
