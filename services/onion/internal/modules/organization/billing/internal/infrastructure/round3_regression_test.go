package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/billing/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestRound3BillingSummaryUsesSubscriptionState(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	ownerID := int64(987654401)

	result, err := db.ExecContext(ctx, `
		INSERT INTO subscriptions (owner_id, plan_id, status, seat_count, next_billing_date)
		SELECT ?, id, 'active', 3, '2030-01-01'
		FROM plans WHERE slug = 'pro' AND is_active = TRUE`, ownerID)
	if err != nil {
		t.Fatalf("seed subscription: %v", err)
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		t.Fatalf("seeded subscription rows = %d, want 1", affected)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM subscriptions WHERE owner_id = ?`, ownerID)
	})

	summary, err := repo.GetSummary(ctx, ownerID)
	if err != nil {
		t.Fatalf("get summary: %v", err)
	}
	if summary.Plan != "pro" || summary.Status != "active" {
		t.Fatalf("summary = %+v, want active pro plan", summary)
	}
	if summary.CurrentUsage != 3 || summary.UsageLimit != 30 {
		t.Fatalf("usage = %v/%v, want 3/30", summary.CurrentUsage, summary.UsageLimit)
	}

	if _, err := repo.ListInvoices(ctx, ownerID); !errors.Is(err, infrastructure.ErrBillingInvoicesUnavailable) {
		t.Fatalf("ListInvoices error = %v, want ErrBillingInvoicesUnavailable", err)
	}
}
