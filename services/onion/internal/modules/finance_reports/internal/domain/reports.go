package domain

type BalanceSheet struct {
	TotalAssets        float64 `json:"total_assets"`
	TotalLiabilities   float64 `json:"total_liabilities"`
	TotalEquity        float64 `json:"total_equity"`
	CashBalance        float64 `json:"cash_balance"`
	AccountsReceivable float64 `json:"accounts_receivable"`
	AccountsPayable    float64 `json:"accounts_payable"`
}

type CashFlow struct {
	Period           string  `json:"period"`
	OperatingInflow  float64 `json:"operating_inflow"`
	OperatingOutflow float64 `json:"operating_outflow"`
	NetOperating     float64 `json:"net_operating"`
	NetCashFlow      float64 `json:"net_cash_flow"`
}
