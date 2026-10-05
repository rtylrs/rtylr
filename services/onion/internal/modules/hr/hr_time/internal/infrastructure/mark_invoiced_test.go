package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func seedApprovedBillableEntry(ctx context.Context, t *testing.T, repo *infrastructure.Repository, description string) int64 {
	t.Helper()
	created, err := repo.CreateTimeEntry(ctx, testutil.OwnerID, domain.TimeEntryRecord{
		EmployeeID:      4101,
		Date:            "2026-04-03",
		DurationMinutes: 90,
		Description:     description,
		IsBillable:      true,
		Status:          "approved",
	})
	if err != nil {
		t.Fatalf("CreateTimeEntry: %v", err)
	}
	return created.ID
}

// The whole batch lands or none of it does: one foreign id must not stop the
// eligible ids from being stamped, and the response's skipped accounting must
// match reality.
func TestMarkEntriesInvoiced_BatchSemantics(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := infrastructure.NewRepository(db)

	first := seedApprovedBillableEntry(ctx, t, repo, "invoiced batch a")
	second := seedApprovedBillableEntry(ctx, t, repo, "invoiced batch b")
	foreign := int64(9_999_999)
	const firstInvoiceID int64 = 41_000_001
	const secondInvoiceID int64 = 41_000_002

	marked, err := repo.MarkEntriesInvoiced(ctx, testutil.OwnerID, firstInvoiceID, []int64{first, second, foreign})
	if err != nil {
		t.Fatalf("MarkEntriesInvoiced: %v", err)
	}
	if marked != 2 {
		t.Fatalf("marked %d entries, want 2 (foreign id counts as skipped)", marked)
	}

	var invoiceID int64
	for _, id := range []int64{first, second} {
		if err := db.QueryRowContext(ctx,
			"SELECT billable_invoice_id FROM hr_time_entries WHERE id = ? AND business_id = ?",
			id, testutil.OwnerID,
		).Scan(&invoiceID); err != nil {
			t.Fatalf("read billable_invoice_id for %d: %v", id, err)
		}
		if invoiceID != firstInvoiceID {
			t.Fatalf("entry %d has billable_invoice_id %d, want %d", id, invoiceID, firstInvoiceID)
		}
	}

	// Idempotent skip: re-marking under a different invoice must not overwrite
	// the first settlement.
	markedAgain, err := repo.MarkEntriesInvoiced(ctx, testutil.OwnerID, secondInvoiceID, []int64{first})
	if err != nil {
		t.Fatalf("re-mark: %v", err)
	}
	if markedAgain != 0 {
		t.Fatalf("re-mark affected %d rows, want 0", markedAgain)
	}
	if err := db.QueryRowContext(ctx,
		"SELECT billable_invoice_id FROM hr_time_entries WHERE id = ? AND business_id = ?",
		first, testutil.OwnerID,
	).Scan(&invoiceID); err != nil {
		t.Fatalf("read billable_invoice_id after re-mark: %v", err)
	}
	if invoiceID != firstInvoiceID {
		t.Fatalf("entry %d was overwritten to %d; already-invoiced rows are skip-only", first, invoiceID)
	}

	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(),
			"DELETE FROM hr_time_entries WHERE id IN (?, ?) AND business_id = ?",
			first, second, testutil.OwnerID,
		)
	})
}

// An empty batch is a no-op rather than an error, so callers can pipe filtered
// id lists straight through without pre-checking length.
func TestMarkEntriesInvoiced_EmptyBatchNoOp(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := infrastructure.NewRepository(db)

	marked, err := repo.MarkEntriesInvoiced(ctx, testutil.OwnerID, 1, nil)
	if err != nil {
		t.Fatalf("MarkEntriesInvoiced empty: %v", err)
	}
	if marked != 0 {
		t.Fatalf("empty batch marked %d rows", marked)
	}
}
