package domain

import (
	"errors"
	"time"

	"github.com/voxire/onion/internal/platform/financejournal"
)

var (
	ErrJournalEntryNotFound      = errors.New("journal entry not found")
	ErrJournalEntryAlreadyPosted = errors.New("journal entry already posted")
	ErrInvalidJournalEntry       = errors.New("invalid journal entry")
)

type JournalEntry struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	Date        time.Time
	Description string
	Status      string
	PostedAt    time.Time
	Lines       []JournalLine
	CreatedAt   time.Time
}

type JournalLine struct {
	ID          int64
	EntryID     int64
	AccountID   int64
	AccountName string
	Debit       float64
	Credit      float64
}

type JournalEntryInput struct {
	Date        string
	Description string
	Lines       []JournalLineInput
}

type JournalLineInput struct {
	AccountID int64
	Debit     float64
	Credit    float64
}

// PayrollJournalInput is the finance-owned representation of a payroll export.
// The repository turns it into one idempotent, posted double-entry record.
type PayrollJournalInput = financejournal.PayrollJournalEntry

// IsPosted reports whether the entry has been posted to the ledger.
func (j *JournalEntry) IsPosted() bool {
	return j.Status == "posted"
}

// IsDraft reports whether the entry is still in draft state and has not been
// posted yet.
func (j *JournalEntry) IsDraft() bool {
	return j.Status == "draft" || j.Status == ""
}

// TotalDebits returns the sum of all debit amounts across the entry's lines.
func (j *JournalEntry) TotalDebits() float64 {
	var total float64
	for _, l := range j.Lines {
		total += l.Debit
	}
	return total
}

// TotalCredits returns the sum of all credit amounts across the entry's lines.
func (j *JournalEntry) TotalCredits() float64 {
	var total float64
	for _, l := range j.Lines {
		total += l.Credit
	}
	return total
}
