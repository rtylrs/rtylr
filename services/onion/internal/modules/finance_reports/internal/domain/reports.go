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

// ProfitAndLoss is a simplified P&L summary for a given period.
type ProfitAndLoss struct {
	Period       string  `json:"period"`
	Revenue      float64 `json:"revenue"`
	Expenses     float64 `json:"expenses"`
	GrossProfit  float64 `json:"gross_profit"`
	NetProfit    float64 `json:"net_profit"`
	ProfitMargin float64 `json:"profit_margin"` // percentage: 0-100
}

// FinancialSummary bundles the most-used single-page KPIs.
type FinancialSummary struct {
	TotalRevenue     float64 `json:"total_revenue"`
	TotalExpenses    float64 `json:"total_expenses"`
	NetProfit        float64 `json:"net_profit"`
	WorkingCapital   float64 `json:"working_capital"`
	CashBalance      float64 `json:"cash_balance"`
	UnpaidReceivable float64 `json:"unpaid_receivable"`
	UnpaidPayable    float64 `json:"unpaid_payable"`
}
