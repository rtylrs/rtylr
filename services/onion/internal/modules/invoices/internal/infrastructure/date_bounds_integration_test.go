package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/testutil"
)

func TestCountInvoiceSequence_UsesExclusiveYearBounds(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	businessID := time.Now().UTC().UnixNano()

	insertInvoiceAt(t, db, businessID, "start", time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC))
	insertInvoiceAt(t, db, businessID, "end", time.Date(2024, time.December, 31, 23, 59, 59, 0, time.UTC))
	insertInvoiceAt(t, db, businessID, "next", time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC))
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM finance_invoices WHERE business_id = ?", businessID)
	})

	queries := sqlc.New(db)
	count, err := queries.CountInvoiceSequence(ctx, sqlc.CountInvoiceSequenceParams{
		BusinessID:     businessID,
		CreatedAtStart: sql.NullTime{Time: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC), Valid: true},
		CreatedAtEnd:   sql.NullTime{Time: time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC), Valid: true},
	})
	if err != nil {
		t.Fatalf("CountInvoiceSequence: %v", err)
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}

	emptyCount, err := queries.CountInvoiceSequence(ctx, sqlc.CountInvoiceSequenceParams{
		BusinessID:     businessID,
		CreatedAtStart: sql.NullTime{Time: time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC), Valid: true},
		CreatedAtEnd:   sql.NullTime{Time: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC), Valid: true},
	})
	if err != nil {
		t.Fatalf("CountInvoiceSequence empty period: %v", err)
	}
	if emptyCount != 0 {
		t.Fatalf("empty count = %d, want 0", emptyCount)
	}
}

func insertInvoiceAt(t *testing.T, db *sql.DB, businessID int64, suffix string, createdAt time.Time) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO finance_invoices
			(owner_id, business_id, invoice_number, client_name, issue_date, due_date, created_at)
		 VALUES (?, ?, ?, 'date bounds test', '2024-01-01', '2024-01-01', ?)`,
		businessID, businessID, fmt.Sprintf("date-bounds-%d-%s", businessID, suffix), createdAt,
	)
	if err != nil {
		t.Fatalf("insert invoice at %s: %v", createdAt, err)
	}
}
