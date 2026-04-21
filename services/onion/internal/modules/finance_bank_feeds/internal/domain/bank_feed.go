package domain

import (
	"errors"
	"time"
)

var ErrBankFeedTransactionNotFound = errors.New("bank feed transaction not found")

// BankFeedTransaction is a single transaction imported from a bank feed.
// Amounts are stored as integer minor units (cents for USD, etc.) — never
// floats — to avoid rounding on money.
type BankFeedTransaction struct {
	ID               int64     `json:"id"`
	OwnerID          int64     `json:"owner_id"`
	BankAccountID    int64     `json:"bank_account_id"`
	AmountMinorUnits int64     `json:"amount_minor_units"`
	Currency         string    `json:"currency"`
	OccurredAt       time.Time `json:"occurred_at"`
	Description      string    `json:"description"`
	ExternalID       string    `json:"external_id"`
	MatchedExpenseID *int64    `json:"matched_expense_id"`
	MatchedInvoiceID *int64    `json:"matched_invoice_id"`
	CreatedAt        time.Time `json:"created_at"`
}
