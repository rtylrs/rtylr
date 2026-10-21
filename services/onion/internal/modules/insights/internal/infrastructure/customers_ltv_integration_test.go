package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/testutil"
)

// TestGetLTVDistribution_CountsPaidSpendOlderThanOneYear pins B19-R05: the
// histogram is served as lifetime value, so paid spend older than the trailing
// 12 months must still land in its value bucket, not in the 0-100 bucket.
func TestGetLTVDistribution_CountsPaidSpendOlderThanOneYear(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	businessID := time.Now().UTC().UnixNano()

	res, err := db.ExecContext(ctx,
		`INSERT INTO crm_customers (owner_id, business_id, first_name, last_name)
		 VALUES (?, ?, 'ltv', 'old-spend')`,
		businessID, businessID)
	if err != nil {
		t.Fatalf("insert customer: %v", err)
	}
	customerID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("customer id: %v", err)
	}

	if _, err := db.ExecContext(ctx,
		`INSERT INTO pos_orders (owner_id, business_id, order_number, customer_id, total, payment_status, created_at)
		 VALUES (?, ?, 'ltv-old-order', ?, 2500.00, 'paid', DATE_SUB(UTC_TIMESTAMP(), INTERVAL 18 MONTH))`,
		businessID, businessID, customerID); err != nil {
		t.Fatalf("insert order: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM pos_orders WHERE customer_id = ?", customerID)
		_, _ = db.ExecContext(ctx, "DELETE FROM crm_customers WHERE id = ?", customerID)
	})

	rows, err := newRepository(db).GetLTVDistribution(ctx, businessID)
	if err != nil {
		t.Fatalf("GetLTVDistribution: %v", err)
	}
	byLabel := make(map[string]int64, len(rows))
	for _, row := range rows {
		byLabel[row.Label] = row.CustomerCount
	}
	if got := byLabel["1k-5k"]; got != 1 {
		t.Errorf("1k-5k bucket = %d customers, want 1 (18-month-old $2500 spend)", got)
	}
	if got := byLabel["0-100"]; got != 0 {
		t.Errorf("0-100 bucket = %d customers, want 0 (lifetime spend is $2500)", got)
	}
}
