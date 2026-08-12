package domain

// Invoice status constants used throughout the invoices module.
const (
	StatusDraft     = "draft"
	StatusSent      = "sent"
	StatusViewed    = "viewed"
	StatusPartial   = "partial"
	StatusPaid      = "paid"
	StatusOverdue   = "overdue"
	StatusCancelled = "cancelled"
	StatusVoid      = "void"
)

// ValidInvoiceStatuses is the complete set of statuses an invoice may hold.
var ValidInvoiceStatuses = map[string]struct{}{
	StatusDraft:     {},
	StatusSent:      {},
	StatusViewed:    {},
	StatusPartial:   {},
	StatusPaid:      {},
	StatusOverdue:   {},
	StatusCancelled: {},
	StatusVoid:      {},
}

// TerminalStatuses are statuses from which an invoice should not be
// further edited (other than voiding).
var TerminalStatuses = map[string]struct{}{
	StatusPaid:      {},
	StatusCancelled: {},
	StatusVoid:      {},
}

// IsTerminal reports whether the given status is terminal.
func IsTerminal(status string) bool {
	_, ok := TerminalStatuses[status]
	return ok
}

// IsValidStatus reports whether the given status is a known invoice status.
func IsValidStatus(status string) bool {
	_, ok := ValidInvoiceStatuses[status]
	return ok
}

// Payment method constants used across the invoices and payments sub-domains.
const (
	PaymentMethodCash         = "cash"
	PaymentMethodCard         = "card"
	PaymentMethodBankTransfer = "bank_transfer"
	PaymentMethodOther        = "other"
)
