package domain

import (
	"time"
)

// BankFeedTransaction is a single transaction imported from a bank feed.
// Amounts are stored as integer minor units (cents for USD, etc.) — never
// floats — to avoid rounding on money.
type BankFeedTransaction struct {
	ID               int64     `json:"id"`
	OwnerID          int64     `json:"owner_id"`
	BusinessID       int64     `json:"business_id"`
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

// IsMatched reports whether the transaction has been matched to an expense or
// an invoice.
func (t *BankFeedTransaction) IsMatched() bool {
	return t.MatchedExpenseID != nil || t.MatchedInvoiceID != nil
}

// AmountDecimal converts the minor-unit amount to a decimal value using the
// given divisor (e.g. 100 for currencies with 2 decimal places).
func (t *BankFeedTransaction) AmountDecimal(divisor int64) float64 {
	if divisor <= 0 {
		divisor = 100
	}
	return float64(t.AmountMinorUnits) / float64(divisor)
}

// IsCredit reports whether the transaction is a credit (positive amount).
func (t *BankFeedTransaction) IsCredit() bool {
	return t.AmountMinorUnits > 0
}

// IsDebit reports whether the transaction is a debit (negative amount).
func (t *BankFeedTransaction) IsDebit() bool {
	return t.AmountMinorUnits < 0
}
