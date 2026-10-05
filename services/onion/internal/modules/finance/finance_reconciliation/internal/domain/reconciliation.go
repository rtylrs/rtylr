package domain

import (
	"errors"
	"time"
)

var (
	ErrBankFeedNotFound = errors.New("bank feed transaction not found")
	ErrTargetNotFound   = errors.New("reconciliation target not found")
	ErrAlreadyMatched   = errors.New("bank feed transaction is already matched")
	ErrInvalidTarget    = errors.New("invalid reconciliation target")
)

const (
	TargetTypeInvoice = "invoice"
	TargetTypeExpense = "expense"
)

// UnmatchedBankTransaction is a bank feed transaction that has not yet been
// reconciled against an invoice or expense.
type UnmatchedBankTransaction struct {
	ID               int64
	BankAccountID    int64
	AmountMinorUnits int64
	Currency         string
	OccurredAt       time.Time
	Description      string
	ExternalID       string
}

// UnmatchedInvoice is an outstanding invoice still awaiting a bank match.
// Amounts are exposed as integer minor units to match the bank-side shape.
type UnmatchedInvoice struct {
	ID               int64
	InvoiceNumber    string
	ClientName       string
	AmountMinorUnits int64
	Currency         string
	IssueDate        time.Time
	DueDate          time.Time
}

// MatchSuggestion pairs an unmatched bank transaction with an unmatched
// invoice (or, later, an expense) ranked by amount + date proximity. The
// score is a positive number where lower is better — it is an L1 distance
// on (amount_minor_units, days_between_dates) with a small weight on the
// date axis so same-amount rows with nearby dates sort first.
type MatchSuggestion struct {
	BankFeedID       int64
	TargetType       string
	TargetID         int64
	AmountMinorUnits int64
	DaysApart        int
	Score            float64
}

// ReconciliationOverview is the payload returned by GET /finance/reconciliation.
type ReconciliationOverview struct {
	UnmatchedBankTxns []UnmatchedBankTransaction
	UnmatchedInvoices []UnmatchedInvoice
	Suggestions       []MatchSuggestion
	BankTxnsTruncated bool
	InvoicesTruncated bool
	ExpensesTruncated bool
	PreviewLimit      int
}

// MatchInput is the payload accepted by POST /finance/reconciliation/match.
type MatchInput struct {
	BankFeedID int64
	TargetType string
	TargetID   int64
	MatchedBy  int64
}

// IsExactAmountMatch reports whether the suggestion has zero amount difference
// (the bank transaction amount equals the target amount exactly).
func (s *MatchSuggestion) IsExactAmountMatch() bool {
	return s.AmountMinorUnits == 0
}

// IsHighConfidence reports whether the suggestion qualifies as high confidence
// based on score threshold. Scores below 500 are considered high confidence.
func (s *MatchSuggestion) IsHighConfidence() bool {
	return s.Score < 500
}

// IsValidTargetType reports whether the target type is one of the known
// reconciliation target types.
func IsValidTargetType(targetType string) bool {
	return targetType == TargetTypeInvoice || targetType == TargetTypeExpense
}

// HasSuggestions reports whether the overview has any match suggestions.
func (o *ReconciliationOverview) HasSuggestions() bool {
	return len(o.Suggestions) > 0
}

// HasUnmatchedItems reports whether there are any unmatched bank transactions
// or unmatched invoices in the overview.
func (o *ReconciliationOverview) HasUnmatchedItems() bool {
	return len(o.UnmatchedBankTxns) > 0 || len(o.UnmatchedInvoices) > 0
}
