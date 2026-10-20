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
	// ExpensesByCategory splits OperatingExpenses by category, largest first.
	ExpensesByCategory []CategoryTotal
	// PendingExpenses totals the period's expenses awaiting approval, which
	// OperatingExpenses leaves out until they are approved.
	PendingExpenses     float64
	PendingExpenseCount int64
}

// CategoryTotal is one category's share of a period's approved expenses.
type CategoryTotal struct {
	// CategoryID is nil for expenses filed under no category.
	CategoryID *int64
	Name       string
	Color      string
	Amount     float64
}

// MonthTotal is one calendar month of revenue and approved expenses.
type MonthTotal struct {
	Month    string // YYYY-MM
	Revenue  float64
	Expenses float64
}

// Activity is one recent money movement: an expense or a paid invoice.
type Activity struct {
	Kind        string // "expense" or "payment"
	ID          int64
	Description string
	Amount      float64
	Date        string // YYYY-MM-DD
	Status      string // the expense status; empty for payments
}

// OverdueInvoice is an unpaid invoice past its due date.
type OverdueInvoice struct {
	ID            int64
	InvoiceNumber string
	ClientName    string
	AmountDue     float64
	DaysOverdue   int64
}

// Dashboard holds the finance dashboard summary.
type Dashboard struct {
	TotalReceivable  float64
	OverdueAmount    float64
	TotalExpensesMTD float64
	TotalRevenueMTD  float64
	NetProfitMTD     float64
	// PendingExpenses totals every expense awaiting approval, whatever its
	// date. None of the figures above include them.
	PendingExpenses     float64
	PendingExpenseCount int64
	Monthly             []MonthTotal
	RecentActivity      []Activity
	OverdueInvoices     []OverdueInvoice
	OverdueInvoiceCount int64
}

// TaxReport holds tax data for a given period.
type TaxReport struct {
	Period       string
	TaxCollected float64
	TaxOwed      float64
}
