package infrastructure

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/customers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/testutil"
)

func TestCountNewCustomersThisMonth_UsesExclusiveBounds(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	businessID := time.Now().UTC().UnixNano()
	monthStart := time.Date(2024, time.December, 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)

	insertCustomerAt(t, db, businessID, monthStart)
	insertCustomerAt(t, db, businessID, monthEnd.Add(-time.Second))
	insertCustomerAt(t, db, businessID, monthEnd)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM crm_customers WHERE business_id = ?", businessID)
	})

	queries := sqlc.New(db)
	count, err := queries.CountNewCustomersThisMonth(ctx, sqlc.CountNewCustomersThisMonthParams{
		BusinessID:     businessID,
		CreatedAtStart: sql.NullTime{Time: monthStart, Valid: true},
		CreatedAtEnd:   sql.NullTime{Time: monthEnd, Valid: true},
	})
	if err != nil {
		t.Fatalf("CountNewCustomersThisMonth: %v", err)
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
}

func insertCustomerAt(t *testing.T, db *sql.DB, businessID int64, createdAt time.Time) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO crm_customers (owner_id, business_id, first_name, created_at)
		 VALUES (?, ?, 'date bounds test', ?)`,
		businessID, businessID, createdAt,
	)
	if err != nil {
		t.Fatalf("insert customer at %s: %v", createdAt, err)
	}
}
