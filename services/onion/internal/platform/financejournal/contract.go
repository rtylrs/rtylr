package financejournal

import "context"

// PayrollJournalEntry is the cross-module contract for posting an approved
// payroll run to the central finance journal.
type PayrollJournalEntry struct {
	OwnerID     int64
	BusinessID  int64
	RunID       int64
	PeriodStart string
	PeriodEnd   string
	TotalGross  float64
	TotalNet    float64
}

// PayrollJournalPoster is implemented by the finance journal module and
// consumed by payroll without exposing persistence details.
type PayrollJournalPoster interface {
	PostPayrollJournal(ctx context.Context, entry PayrollJournalEntry) error
}
