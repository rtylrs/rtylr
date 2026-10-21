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
