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

// MoneyDerivedStatuses are computed from the recorded payments, never set by a
// caller. Only RecordPayment may write them: "paid" adds total to revenue and
// "partial" asserts cash was collected, so a status-only write would report
// money that was never received.
var MoneyDerivedStatuses = map[string]struct{}{
	StatusPartial: {},
	StatusPaid:    {},
}

// ManuallyAssignableStatuses are the statuses a status-only write may set:
// the ENUM minus the money-derived pair. See ValidateSettableInvoiceStatus.
var ManuallyAssignableStatuses = map[string]struct{}{
	StatusDraft:     {},
	StatusSent:      {},
	StatusViewed:    {},
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

// IsManuallyAssignableStatus reports whether a status-only write may set the
// given status, i.e. whether it is known and not derived from payments.
func IsManuallyAssignableStatus(status string) bool {
	_, ok := ManuallyAssignableStatuses[status]
	return ok
}

// Payment method constants used across the invoices and payments sub-domains.
const (
	PaymentMethodCash         = "cash"
	PaymentMethodCard         = "card"
	PaymentMethodBankTransfer = "bank_transfer"
	PaymentMethodOther        = "other"
)
