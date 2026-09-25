package domain

// IsTerminal reports whether the purchase order is in a terminal status
// (received or cancelled) and can no longer be modified.
func (po PurchaseOrder) IsTerminal() bool {
	return po.Status == POStatusReceived || po.Status == POStatusCancelled
}

// IsOpen reports whether the purchase order is in an open (non-terminal) status.
func (po PurchaseOrder) IsOpen() bool {
	return !po.IsTerminal()
}

// StatusLabel returns a human-friendly label for the purchase order status.
func (po PurchaseOrder) StatusLabel() string {
	switch po.Status {
	case POStatusDraft:
		return "Draft"
	case POStatusSent:
		return "Sent to Supplier"
	case POStatusApproved:
		return "Approved"
	case POStatusPartial:
		return "Partially Received"
	case POStatusReceived:
		return "Fully Received"
	case POStatusCancelled:
		return "Cancelled"
	default:
		return string(po.Status)
	}
}
