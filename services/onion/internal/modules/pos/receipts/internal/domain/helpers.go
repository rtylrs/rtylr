package domain

import (
	"fmt"
	"time"
)

// FormatReceiptHeader produces a compact header line for receipt printing.
func (r Receipt) FormatReceiptHeader() string {
	return fmt.Sprintf("Order: %s | %s", r.OrderNumber, r.PaidAt.Format(time.RFC822))
}

// FormatTotal returns the total as a currency-formatted string.
func (r Receipt) FormatTotal() string {
	return fmt.Sprintf("%.2f", r.Total)
}

// IsToday reports whether the receipt was issued on the given date (UTC).
func (r Receipt) IsToday(now time.Time) bool {
	rY, rM, rD := r.PaidAt.UTC().Date()
	nY, nM, nD := now.UTC().Date()
	return rY == nY && rM == nM && rD == nD
}
