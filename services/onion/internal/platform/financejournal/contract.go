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

// LedgerSyncer posts approved expenses and recorded invoice payments that are
// not in the journal yet. Each source posts at most once, so it is safe to
// call before every ledger read; reports call it so the books they show
// include everything the operational screens already do.
type LedgerSyncer interface {
	SyncOperationalPostings(ctx context.Context, businessID int64) error
}
