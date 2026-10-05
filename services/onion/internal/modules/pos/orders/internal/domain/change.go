package domain

import "fmt"

// ChangeResult holds the result of a cash payment calculation.
type ChangeResult struct {
	AmountDue  float64
	AmountPaid float64
	Change     float64
}

// ComputeChange calculates the change owed to the customer.
// It returns an error if amountPaid is less than total.
func ComputeChange(total, amountPaid float64) (ChangeResult, error) {
	if amountPaid < total {
		return ChangeResult{}, fmt.Errorf("%w: amount paid (%.2f), total (%.2f)", ErrPaymentInsufficient, amountPaid, total)
	}
	return ChangeResult{
		AmountDue:  total,
		AmountPaid: amountPaid,
		Change:     amountPaid - total,
	}, nil
}
