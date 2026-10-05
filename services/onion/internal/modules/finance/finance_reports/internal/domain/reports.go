package domain

type BalanceSheet struct {
	TotalAssets        float64
	TotalLiabilities   float64
	TotalEquity        float64
	CashBalance        float64
	AccountsReceivable float64
	AccountsPayable    float64
}

type CashFlow struct {
	Period           string
	OperatingInflow  float64
	OperatingOutflow float64
	NetOperating     float64
	NetCashFlow      float64
}
