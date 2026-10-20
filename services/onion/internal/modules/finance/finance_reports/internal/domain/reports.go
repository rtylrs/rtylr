package domain

type BalanceSheet struct {
	// AsOf is the report date, YYYY-MM-DD.
	AsOf  string
	Lines []BalanceSheetLine
	// NetIncome is revenue less expenses posted by AsOf; it is in TotalEquity.
	NetIncome          float64
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

// BalanceSheetLine is one ledger account's balance on the report date.
type BalanceSheetLine struct {
	AccountID int64
	Code      string
	Name      string
	Type      string
	Amount    float64
}
