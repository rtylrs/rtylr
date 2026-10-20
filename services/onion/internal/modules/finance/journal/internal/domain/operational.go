package domain

import "github.com/voxire/onion/internal/platform/money"

// Source types for journal entries posted from operational records.
const (
	SourceExpense        = "expense"
	SourceInvoicePayment = "invoice_payment"
)

// OperationalSource is an approved expense or a recorded invoice payment that
// is not in the journal yet.
type OperationalSource struct {
	Kind        string
	ID          int64
	OwnerID     int64
	BusinessID  int64
	Date        string
	Description string
	Amount      money.Amount
	Currency    string
	// PaidInCash routes the money side to Cash instead of Bank.
	PaidInCash bool
}
