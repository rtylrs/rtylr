package invoicenumber

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/voxire/onion/internal/modules/finance/internal/invoicenumber/persistence/sqlc"
)

// Allocate participates in the caller's Finance transaction. Rollback does not consume a number.
func Allocate(ctx context.Context, tx *sql.Tx, businessID int64, year int) (string, error) {
	q := sqlc.New(tx)
	prefix := fmt.Sprintf("INV-%d-", year)
	if err := q.EnsureInvoiceSequence(ctx, sqlc.EnsureInvoiceSequenceParams{BusinessID: businessID, Prefix: prefix}); err != nil {
		return "", err
	}
	sequence, err := q.LockInvoiceSequence(ctx, sqlc.LockInvoiceSequenceParams{BusinessID: businessID, Prefix: prefix})
	if err != nil {
		return "", err
	}
	existing, err := q.MaxInvoiceSequence(ctx, sqlc.MaxInvoiceSequenceParams{BusinessID: businessID, Prefix: prefix})
	if err != nil {
		return "", err
	}
	if existing > sequence {
		sequence = existing
	}
	if err := q.AdvanceInvoiceSequence(ctx, sqlc.AdvanceInvoiceSequenceParams{BusinessID: businessID, Prefix: prefix, LastSequence: sequence + 1}); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%04d", prefix, sequence+1), nil
}
