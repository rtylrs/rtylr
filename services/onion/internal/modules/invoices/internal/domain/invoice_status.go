package domain

// Invoice status lifecycle for status-only writes (the PATCH /invoices/{id}/status
// endpoint, the legacy /finance/invoices/status handler and any future caller of
// Service.UpdateInvoiceStatus). RecordPayment is the only writer of the
// money-derived statuses and is deliberately absent from this table.
//
// The table is keyed by the current status and holds the statuses a status-only
// write may move to. Everything not listed is refused.
//
//	current   -> draft sent viewed overdue cancelled void
//	draft       yes  yes  no     no      yes       yes
//	sent        no   yes  yes    yes     yes       yes
//	viewed      no   yes  yes    yes     yes       yes
//	partial     no   no   no     yes     yes       yes
//	overdue     no   yes  yes    yes     yes       yes
//	paid        no   no   no     no      no        no
//	cancelled   no   no   no     no      no        no
//	void        no   no   no     no      no        no
//
// Reading the rows:
//
//   - draft may only be issued (sent) or closed. A draft that was never sent
//     has not been delivered, so it cannot be viewed or overdue.
//   - sent, viewed and overdue are the same open receivable wearing a
//     different delivery or ageing label, so they may be relabelled freely.
//   - partial has cash against it, so it may not be relabelled back to a
//     status that reads as "nothing collected yet". Closing it out (cancelled,
//     void) is still allowed: this module has no refund or credit-note flow, so
//     refusing it here would leave no way to close a partially paid invoice.
//   - the terminal rows are empty and must stay in step with TerminalStatuses.
var invoiceStatusTransitions = map[string]map[string]struct{}{
	StatusDraft: {
		StatusDraft:     {},
		StatusSent:      {},
		StatusCancelled: {},
		StatusVoid:      {},
	},
	StatusSent: {
		StatusSent:      {},
		StatusViewed:    {},
		StatusOverdue:   {},
		StatusCancelled: {},
		StatusVoid:      {},
	},
	StatusViewed: {
		StatusSent:      {},
		StatusViewed:    {},
		StatusOverdue:   {},
		StatusCancelled: {},
		StatusVoid:      {},
	},
	StatusPartial: {
		StatusOverdue:   {},
		StatusCancelled: {},
		StatusVoid:      {},
	},
	StatusOverdue: {
		StatusSent:      {},
		StatusViewed:    {},
		StatusOverdue:   {},
		StatusCancelled: {},
		StatusVoid:      {},
	},
	StatusPaid:      {},
	StatusCancelled: {},
	StatusVoid:      {},
}

// ValidateSettableInvoiceStatus reports whether status may be written by a
// status-only write at all, independently of the invoice's current status.
// Call it before reading the invoice so that a money-derived or unknown status
// is refused as a bad request rather than as a lifecycle conflict.
func ValidateSettableInvoiceStatus(status string) error {
	if !IsValidStatus(status) {
		return ErrInvalidInvoiceStatus
	}
	if !IsManuallyAssignableStatus(status) {
		return ErrInvoiceStatusNotSettable
	}
	return nil
}

// CanTransitionInvoiceStatus reports whether an invoice may be moved from one
// status to another by a status-only write.
func CanTransitionInvoiceStatus(from, to string) error {
	if err := ValidateSettableInvoiceStatus(to); err != nil {
		return err
	}
	allowed, known := invoiceStatusTransitions[from]
	if !known {
		return ErrInvoiceNotDraft
	}
	if _, ok := allowed[to]; !ok {
		return ErrInvoiceNotDraft
	}
	return nil
}
