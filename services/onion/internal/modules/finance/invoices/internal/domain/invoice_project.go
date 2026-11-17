package domain

import "errors"

// ErrInvoiceProjectLocked reports a project tag change on an invoice that is
// cancelled or void.
var ErrInvoiceProjectLocked = errors.New("cancelled and void invoices cannot change project")

// projectTagLockedStatuses are the statuses whose Work project tag is frozen.
//
// The tag is reporting metadata, not part of the billed document, so it may
// change in any open or settled state, paid included: a paid invoice is
// exactly what a project's collected total counts. Cancelled and void are
// different. Both are dead ends in the status lifecycle (see
// invoiceStatusTransitions), both mean the invoice was never billed and never
// will be, and the per-project dashboard leaves both out of every total, so a
// re-tag there would change nothing a reader can see while still rewriting a
// closed record.
var projectTagLockedStatuses = map[string]struct{}{
	StatusCancelled: {},
	StatusVoid:      {},
}

// CanChangeInvoiceProject reports whether an invoice in the given status may
// have its Work project tag set or cleared.
func CanChangeInvoiceProject(status string) error {
	if !IsValidStatus(status) {
		return ErrInvalidInvoiceStatus
	}
	if _, locked := projectTagLockedStatuses[status]; locked {
		return ErrInvoiceProjectLocked
	}
	return nil
}
