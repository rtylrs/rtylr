package finance

// ExpenseCategoryTotal is the approved expense spend booked to one category.
// CategoryID is zero for expenses without a category.
type ExpenseCategoryTotal struct {
	CategoryID       int64
	CategoryName     string
	AmountMinorUnits int64
}

// InvoiceAgingBucket is the outstanding receivable in one due-date age band.
type InvoiceAgingBucket struct {
	Bucket           string
	InvoiceCount     int64
	AmountMinorUnits int64
}

// ProjectInvoiceTotals are the invoices tagged with one Work project, in minor
// units. Drafts, cancelled and void invoices do not count. Outstanding is what
// is still due on sent, viewed, partially paid and overdue invoices; Overdue is
// the part of it on invoices marked overdue or past their due date.
type ProjectInvoiceTotals struct {
	InvoiceCount int64
	Invoiced     int64
	Paid         int64
	Outstanding  int64
	Overdue      int64
	OverdueCount int64
}
