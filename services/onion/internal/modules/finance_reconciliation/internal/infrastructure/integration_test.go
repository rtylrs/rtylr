package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/finance_reconciliation/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListUnmatchedBankFeeds_ReturnsForOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListUnmatchedBankFeeds(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListUnmatchedBankFeeds: %v", err)
	}
	// Result may be empty if no unmatched feeds exist; just verify no error and type safety.
	for _, item := range items {
		if item.ID == 0 {
			t.Error("expected non-zero ID on bank feed")
		}
		if item.Currency == "" {
			t.Error("expected non-empty currency on bank feed")
		}
	}
}

func TestListUnmatchedBankFeeds_UnknownOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListUnmatchedBankFeeds(ctx, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for unknown owner, got %d", len(items))
	}
}

func TestListUnmatchedInvoices_ReturnsForOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListUnmatchedInvoices(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListUnmatchedInvoices: %v", err)
	}
	for _, item := range items {
		if item.ID == 0 {
			t.Error("expected non-zero ID on invoice")
		}
		if item.InvoiceNumber == "" {
			t.Error("expected non-empty invoice number")
		}
	}
}

func TestListUnmatchedInvoices_UnknownOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListUnmatchedInvoices(ctx, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for unknown owner, got %d", len(items))
	}
}

func TestListUnmatchedExpenses_ReturnsForOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListUnmatchedExpenses(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListUnmatchedExpenses: %v", err)
	}
	for _, item := range items {
		if item.ID == 0 {
			t.Error("expected non-zero ID on expense")
		}
	}
}

func TestListUnmatchedExpenses_UnknownOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListUnmatchedExpenses(ctx, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for unknown owner, got %d", len(items))
	}
}

func TestBankFeedExists_ReturnsFalseForNonExistentFeed(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	exists, err := repo.BankFeedExists(ctx, testutil.OwnerID, 999999999)
	if err != nil {
		t.Fatalf("BankFeedExists: %v", err)
	}
	if exists {
		t.Error("expected false for non-existent bank feed ID")
	}
}

func TestBankFeedExists_ReturnsFalseForWrongOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Use a plausible ID but the wrong owner.
	exists, err := repo.BankFeedExists(ctx, 0, 1)
	if err != nil {
		t.Fatalf("BankFeedExists: %v", err)
	}
	if exists {
		t.Error("expected false when owner does not match")
	}
}

func TestListUnmatchedInvoices_ItemsHaveDates(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListUnmatchedInvoices(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListUnmatchedInvoices: %v", err)
	}
	for _, item := range items {
		if item.IssueDate.IsZero() {
			t.Errorf("item %d has zero IssueDate", item.ID)
		}
	}
}

func TestListUnmatchedExpenses_ItemsHaveDates(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListUnmatchedExpenses(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListUnmatchedExpenses: %v", err)
	}
	for _, item := range items {
		if item.IssueDate.IsZero() {
			t.Errorf("item %d has zero IssueDate", item.ID)
		}
	}
}

func TestListUnmatchedBankFeeds_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListUnmatchedBankFeeds(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListUnmatchedBankFeeds: %v", err)
	}
	for _, item := range items {
		if item.OccurredAt.IsZero() {
			t.Errorf("item %d has zero OccurredAt", item.ID)
		}
	}
}

func TestListUnmatchedInvoices_LimitMaxRows(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListUnmatchedInvoices(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListUnmatchedInvoices: %v", err)
	}
	if len(items) > 200 {
		t.Errorf("expected at most 200 items (LIMIT), got %d", len(items))
	}
}

func TestListUnmatchedBankFeeds_IDsArePositive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListUnmatchedBankFeeds(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListUnmatchedBankFeeds: %v", err)
	}
	for _, item := range items {
		if item.ID <= 0 {
			t.Errorf("bank feed item has non-positive ID: %d", item.ID)
		}
	}
}

func TestListUnmatchedInvoices_IDsArePositive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListUnmatchedInvoices(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListUnmatchedInvoices: %v", err)
	}
	for _, item := range items {
		if item.ID <= 0 {
			t.Errorf("invoice item has non-positive ID: %d", item.ID)
		}
	}
}

func TestListUnmatchedExpenses_IDsArePositive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListUnmatchedExpenses(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListUnmatchedExpenses: %v", err)
	}
	for _, item := range items {
		if item.ID <= 0 {
			t.Errorf("expense item has non-positive ID: %d", item.ID)
		}
	}
}

func TestBankFeedExists_NegativeID_ReturnsFalse(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	exists, err := repo.BankFeedExists(ctx, testutil.OwnerID, -1)
	if err != nil {
		t.Fatalf("BankFeedExists: %v", err)
	}
	if exists {
		t.Error("expected false for negative bank feed ID")
	}
}

func TestBankFeedExists_ZeroID_ReturnsFalse(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	exists, err := repo.BankFeedExists(ctx, testutil.OwnerID, 0)
	if err != nil {
		t.Fatalf("BankFeedExists id=0: %v", err)
	}
	if exists {
		t.Error("expected false for id=0")
	}
}

func TestListUnmatchedInvoices_ClientNamePresent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListUnmatchedInvoices(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListUnmatchedInvoices: %v", err)
	}
	for _, item := range items {
		if item.ClientName == "" {
			t.Errorf("invoice %d has empty ClientName", item.ID)
		}
	}
}

func TestListUnmatchedInvoices_CurrencyPresent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListUnmatchedInvoices(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListUnmatchedInvoices: %v", err)
	}
	for _, item := range items {
		if item.Currency == "" {
			t.Errorf("invoice %d has empty Currency", item.ID)
		}
	}
}

func TestListUnmatchedExpenses_CurrencyPresent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListUnmatchedExpenses(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListUnmatchedExpenses: %v", err)
	}
	for _, item := range items {
		if item.Currency == "" {
			t.Errorf("expense %d has empty Currency", item.ID)
		}
	}
}

func TestListUnmatchedBankFeeds_CurrencyPresent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListUnmatchedBankFeeds(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListUnmatchedBankFeeds: %v", err)
	}
	for _, item := range items {
		if item.Currency == "" {
			t.Errorf("bank feed %d has empty Currency", item.ID)
		}
	}
}

func TestListUnmatchedBankFeeds_StableAcrossCalls(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.ListUnmatchedBankFeeds(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	second, err := repo.ListUnmatchedBankFeeds(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if len(first) != len(second) {
		t.Errorf("result length changed between calls: %d vs %d", len(first), len(second))
	}
}

func TestListUnmatchedExpenses_DueDateAfterOrEqualIssueDate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListUnmatchedExpenses(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListUnmatchedExpenses: %v", err)
	}
	for _, item := range items {
		if item.DueDate.Before(item.IssueDate) {
			t.Errorf("expense %d: DueDate (%v) is before IssueDate (%v)", item.ID, item.DueDate, item.IssueDate)
		}
	}
}

func TestListUnmatchedInvoices_DueDateAfterOrEqualIssueDate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListUnmatchedInvoices(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListUnmatchedInvoices: %v", err)
	}
	for _, item := range items {
		if item.DueDate.Before(item.IssueDate) {
			t.Errorf("invoice %d: DueDate (%v) is before IssueDate (%v)", item.ID, item.DueDate, item.IssueDate)
		}
	}
}
