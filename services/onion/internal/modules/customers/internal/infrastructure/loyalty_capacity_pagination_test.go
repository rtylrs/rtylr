package infrastructure_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/customers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestApplyLoyaltyDeltaRejectsOversizedLedgerAmountBeforeDatabaseAccess(t *testing.T) {
	repo := &infrastructure.Repository{}
	points := int(domain.MaxLoyaltyPoints + 1)

	err := repo.ApplyLoyaltyDelta(context.Background(), domain.LoyaltyTransaction{
		CustomerID: 1,
		Type:       "redeem",
		Points:     points,
	}, -points, func(int) string { return domain.TierBronze }, nil)
	if !errors.Is(err, domain.ErrLoyaltyPointsOutOfRange) {
		t.Fatalf("error = %v, want ErrLoyaltyPointsOutOfRange", err)
	}
}

func TestApplyLoyaltyDeltaRejectsCumulativeIntCapacityBeforeWriting(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	businessID := int64(700000000 + time.Now().UnixNano()%100000000)

	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM crm_loyalty_transactions WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM crm_customers WHERE business_id = ?", businessID)
	})

	customerID, err := repo.CreateCustomer(ctx, domain.CustomerInput{
		OwnerID:    businessID,
		BusinessID: businessID,
		FirstName:  "Capacity",
		Consent:    []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("CreateCustomer: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		"UPDATE crm_customers SET loyalty_points = ? WHERE id = ? AND business_id = ?",
		domain.MaxLoyaltyPoints, customerID, businessID,
	); err != nil {
		t.Fatalf("seed loyalty balance: %v", err)
	}

	committed := false
	err = repo.ApplyLoyaltyDelta(ctx, domain.LoyaltyTransaction{
		OwnerID:    businessID,
		BusinessID: businessID,
		CustomerID: customerID,
		Type:       "earn",
		Points:     1,
	}, 1, func(int) string { return domain.TierPlatinum }, func() { committed = true })
	if !errors.Is(err, domain.ErrLoyaltyBalanceCapacity) {
		t.Fatalf("error = %v, want ErrLoyaltyBalanceCapacity", err)
	}
	if committed {
		t.Fatal("capacity failure invoked the post-commit callback")
	}

	var balance int
	if err := db.QueryRowContext(ctx,
		"SELECT loyalty_points FROM crm_customers WHERE id = ? AND business_id = ?",
		customerID, businessID,
	).Scan(&balance); err != nil {
		t.Fatalf("read balance: %v", err)
	}
	if int64(balance) != domain.MaxLoyaltyPoints {
		t.Fatalf("balance = %d, want %d", balance, domain.MaxLoyaltyPoints)
	}

	var transactions int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM crm_loyalty_transactions WHERE business_id = ? AND customer_id = ?",
		businessID, customerID,
	).Scan(&transactions); err != nil {
		t.Fatalf("count transactions: %v", err)
	}
	if transactions != 0 {
		t.Fatalf("capacity failure wrote %d transaction rows, want 0", transactions)
	}
}

func TestListLoyaltyTiersPageReturnsStablePagesAndTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	businessID := int64(800000000 + time.Now().UnixNano()%100000000)

	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM crm_loyalty_tiers WHERE business_id = ?", businessID)
	})

	for i := 0; i < 101; i++ {
		_, err := repo.CreateLoyaltyTier(ctx, domain.LoyaltyTierInput{
			OwnerID:    businessID,
			BusinessID: businessID,
			Name:       fmt.Sprintf("Page tier %03d", i),
			Perks:      []byte(`[]`),
			SortOrder:  i,
		})
		if err != nil {
			t.Fatalf("CreateLoyaltyTier(%d): %v", i, err)
		}
	}

	first, total, err := repo.ListLoyaltyTiersPage(ctx, businessID, 1, 100)
	if err != nil {
		t.Fatalf("page 1: %v", err)
	}
	if total != 101 || len(first) != 100 {
		t.Fatalf("page 1 total/len = %d/%d, want 101/100", total, len(first))
	}

	second, total, err := repo.ListLoyaltyTiersPage(ctx, businessID, 2, 100)
	if err != nil {
		t.Fatalf("page 2: %v", err)
	}
	if total != 101 || len(second) != 1 {
		t.Fatalf("page 2 total/len = %d/%d, want 101/1", total, len(second))
	}
	if second[0].Name != "Page tier 100" {
		t.Fatalf("page 2 first tier = %q, want %q", second[0].Name, "Page tier 100")
	}

	all, err := repo.ListLoyaltyTiers(ctx, businessID)
	if err != nil {
		t.Fatalf("complete list: %v", err)
	}
	if len(all) != 101 {
		t.Fatalf("complete list length = %d, want 101", len(all))
	}
}
