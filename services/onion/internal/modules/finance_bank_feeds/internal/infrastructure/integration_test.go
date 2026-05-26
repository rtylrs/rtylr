package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/finance_bank_feeds/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListBankFeedTransactions_ReturnsSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, total, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
	_ = items
}

func TestListBankFeedTransactions_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, tx := range items {
		if tx.OwnerID != testutil.OwnerID {
			t.Errorf("expected OwnerID=%d, got %d for tx id=%d", testutil.OwnerID, tx.OwnerID, tx.ID)
		}
	}
}

func TestListBankFeedTransactions_IsolatedByOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.List(ctx, 9999999, 100, 0)
	if err != nil {
		t.Fatalf("List other owner: %v", err)
	}
	for _, tx := range items {
		if tx.OwnerID == testutil.OwnerID {
			t.Errorf("isolation violated: tx id=%d belongs to OwnerID=%d", tx.ID, tx.OwnerID)
		}
	}
}

func TestListBankFeedTransactions_CountMatchesTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, total, err := repo.List(ctx, testutil.OwnerID, 1000, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if int64(len(items)) > total {
		t.Errorf("rows=%d > total=%d: inconsistent", len(items), total)
	}
}

func TestListBankFeedTransactions_PageSize(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.List(ctx, testutil.OwnerID, 2, 0)
	if err != nil {
		t.Fatalf("List with limit 2: %v", err)
	}
	if len(items) > 2 {
		t.Errorf("expected at most 2 items, got %d", len(items))
	}
}

func TestListBankFeedTransactions_OffsetBeyondTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.List(ctx, testutil.OwnerID, 100, 1000000)
	if err != nil {
		t.Fatalf("List with large offset: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items past end, got %d", len(items))
	}
}

func TestListBankFeedTransactions_CurrencyNonEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, tx := range items {
		if tx.Currency == "" {
			t.Errorf("transaction id=%d has empty Currency", tx.ID)
		}
	}
}

func TestListBankFeedTransactions_IDsArePositive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, tx := range items {
		if tx.ID <= 0 {
			t.Errorf("expected positive ID, got %d", tx.ID)
		}
		if tx.BankAccountID <= 0 {
			t.Errorf("expected positive BankAccountID, got %d for tx id=%d", tx.BankAccountID, tx.ID)
		}
	}
}

func TestListBankFeedTransactions_OccurredAtNonZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, tx := range items {
		if tx.OccurredAt.IsZero() {
			t.Errorf("transaction id=%d has zero OccurredAt", tx.ID)
		}
	}
}

func TestListBankFeedTransactions_ExternalIDPresent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, tx := range items {
		if tx.ExternalID == "" {
			t.Errorf("transaction id=%d has empty ExternalID", tx.ID)
		}
	}
}

func TestListBankFeedTransactions_StableAcrossMultipleCalls(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, total1, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("first List: %v", err)
	}
	_, total2, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("second List: %v", err)
	}
	if total1 != total2 {
		t.Errorf("total changed between calls: %d vs %d", total1, total2)
	}
}
