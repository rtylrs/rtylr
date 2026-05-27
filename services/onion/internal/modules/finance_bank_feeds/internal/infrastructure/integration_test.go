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

func TestListBankFeedTransactions_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, total, err := repo.List(ctx, 0, 100, 0)
	if err != nil {
		t.Fatalf("List owner_id=0: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for owner_id=0, got %d", len(items))
	}
	if total != 0 {
		t.Errorf("expected total=0 for owner_id=0, got %d", total)
	}
}

func TestListBankFeedTransactions_LargeLimit_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.List(ctx, testutil.OwnerID, 10000, 0)
	if err != nil {
		t.Fatalf("List with large limit: %v", err)
	}
	_ = items
}

func TestListBankFeedTransactions_OffsetOne_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.List(ctx, testutil.OwnerID, 100, 1)
	if err != nil {
		t.Fatalf("List offset=1: %v", err)
	}
	_ = items
}

func TestListBankFeedTransactions_PageOneVsPageTwo_Disjoint(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	page1, total, err := repo.List(ctx, testutil.OwnerID, 3, 0)
	if err != nil {
		t.Fatalf("List page1: %v", err)
	}
	if total < 6 {
		t.Skip("fewer than 6 bank feed transactions in fixture, cannot test disjoint pages")
	}

	page2, _, err := repo.List(ctx, testutil.OwnerID, 3, 3)
	if err != nil {
		t.Fatalf("List page2: %v", err)
	}

	ids1 := map[int64]bool{}
	for _, tx := range page1 {
		ids1[tx.ID] = true
	}
	for _, tx := range page2 {
		if ids1[tx.ID] {
			t.Errorf("tx id=%d appears in both page1 and page2", tx.ID)
		}
	}
}

func TestListBankFeedTransactions_TotalGteLen(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, total, err := repo.List(ctx, testutil.OwnerID, 5, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total < int64(len(items)) {
		t.Errorf("total=%d is less than len(items)=%d", total, len(items))
	}
}

func TestListBankFeedTransactions_AmountField(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, tx := range items {
		// AmountMinorUnits is scanned -- a zero value is valid but the field must be readable.
		_ = tx.AmountMinorUnits
	}
}

func TestListBankFeedTransactions_DescriptionField(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, tx := range items {
		_ = tx.Description
	}
}

func TestListBankFeedTransactions_Limit1_AtMostOne(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.List(ctx, testutil.OwnerID, 1, 0)
	if err != nil {
		t.Fatalf("List limit=1: %v", err)
	}
	if len(items) > 1 {
		t.Errorf("expected at most 1 item with limit=1, got %d", len(items))
	}
}

func TestListBankFeedTransactions_UnknownOwner_TotalIsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, total, err := repo.List(ctx, 9999999, 100, 0)
	if err != nil {
		t.Fatalf("List unknown owner: %v", err)
	}
	if total != 0 {
		t.Errorf("expected total=0 for unknown owner, got %d", total)
	}
}

func TestListBankFeedTransactions_NegativeOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, total, err := repo.List(ctx, -1, 100, 0)
	if err != nil {
		t.Fatalf("List owner_id=-1: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for negative owner, got %d", len(items))
	}
	if total != 0 {
		t.Errorf("expected total=0 for negative owner, got %d", total)
	}
}

func TestListBankFeedTransactions_ExternalIDUnique(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	seen := map[string]bool{}
	for _, tx := range items {
		if tx.ExternalID != "" && seen[tx.ExternalID] {
			t.Errorf("duplicate ExternalID=%q found", tx.ExternalID)
		}
		seen[tx.ExternalID] = true
	}
}

func TestListBankFeedTransactions_BankAccountIDPositive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, tx := range items {
		if tx.BankAccountID <= 0 {
			t.Errorf("tx id=%d has non-positive BankAccountID=%d", tx.ID, tx.BankAccountID)
		}
	}
}
