package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/finance_reconciliation/internal/domain"
	"github.com/voxire/onion/internal/modules/finance_reconciliation/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestUnmatchedLists_RespectBusinessScopedMatches(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	businessID := time.Now().UnixNano()
	otherBusinessID := businessID + 1
	insert := func(query string, args ...any) int64 {
		result, err := db.ExecContext(ctx, query, args...)
		if err != nil {
			t.Fatalf("insert fixture: %v", err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("fixture ID: %v", err)
		}
		return id
	}
	matchedFeedID := insert(
		"INSERT INTO finance_bank_feeds (owner_id, business_id, bank_account_id, amount_minor_units, occurred_at, external_id) VALUES (?, ?, ?, ?, NOW(), ?)",
		businessID, businessID, 1, 1000, fmt.Sprintf("matched-feed-%d", businessID),
	)
	unmatchedFeedID := insert(
		"INSERT INTO finance_bank_feeds (owner_id, business_id, bank_account_id, amount_minor_units, occurred_at, external_id) VALUES (?, ?, ?, ?, NOW(), ?)",
		businessID, businessID, 1, 2000, fmt.Sprintf("unmatched-feed-%d", businessID),
	)
	matchedInvoiceID := insert(
		"INSERT INTO finance_invoices (owner_id, business_id, invoice_number, client_name, issue_date, due_date, total) VALUES (?, ?, ?, ?, CURDATE(), CURDATE(), ?)",
		businessID, businessID, fmt.Sprintf("MATCHED-%d", businessID), "Matched client", 10,
	)
	unmatchedInvoiceID := insert(
		"INSERT INTO finance_invoices (owner_id, business_id, invoice_number, client_name, issue_date, due_date, total) VALUES (?, ?, ?, ?, CURDATE(), CURDATE(), ?)",
		businessID, businessID, fmt.Sprintf("UNMATCHED-%d", businessID), "Unmatched client", 20,
	)
	matchedExpenseID := insert(
		"INSERT INTO finance_expenses (owner_id, business_id, description, amount, expense_date) VALUES (?, ?, ?, ?, CURDATE())",
		businessID, businessID, "Matched expense", 10,
	)
	unmatchedExpenseID := insert(
		"INSERT INTO finance_expenses (owner_id, business_id, description, amount, expense_date) VALUES (?, ?, ?, ?, CURDATE())",
		businessID, businessID, "Unmatched expense", 20,
	)
	insert(
		"INSERT INTO finance_reconciliation_matches (owner_id, business_id, bank_feed_id, target_type, target_id) VALUES (?, ?, ?, 'invoice', ?)",
		businessID, businessID, matchedFeedID, matchedInvoiceID,
	)
	insert(
		"INSERT INTO finance_reconciliation_matches (owner_id, business_id, bank_feed_id, target_type, target_id) VALUES (?, ?, ?, 'expense', ?)",
		businessID, businessID, unmatchedFeedID+999999, matchedExpenseID,
	)
	// A match in another tenant must not hide this tenant's unmatched invoice.
	insert(
		"INSERT INTO finance_reconciliation_matches (owner_id, business_id, bank_feed_id, target_type, target_id) VALUES (?, ?, ?, 'invoice', ?)",
		otherBusinessID, otherBusinessID, unmatchedFeedID+888888, unmatchedInvoiceID,
	)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM finance_reconciliation_matches WHERE business_id IN (?, ?)", businessID, otherBusinessID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM finance_bank_feeds WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM finance_invoices WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM finance_expenses WHERE business_id = ?", businessID)
	})

	feeds, err := repo.ListUnmatchedBankFeeds(ctx, businessID)
	if err != nil {
		t.Fatalf("ListUnmatchedBankFeeds: %v", err)
	}
	if containsBankFeed(feeds, matchedFeedID) || !containsBankFeed(feeds, unmatchedFeedID) {
		t.Fatalf("unmatched feeds did not preserve match isolation: %+v", feeds)
	}
	invoices, err := repo.ListUnmatchedInvoices(ctx, businessID)
	if err != nil {
		t.Fatalf("ListUnmatchedInvoices: %v", err)
	}
	if containsUnmatchedInvoice(invoices, matchedInvoiceID) || !containsUnmatchedInvoice(invoices, unmatchedInvoiceID) {
		t.Fatalf("unmatched invoices did not preserve match isolation: %+v", invoices)
	}
	expenses, err := repo.ListUnmatchedExpenses(ctx, businessID)
	if err != nil {
		t.Fatalf("ListUnmatchedExpenses: %v", err)
	}
	if containsUnmatchedInvoice(expenses, matchedExpenseID) || !containsUnmatchedInvoice(expenses, unmatchedExpenseID) {
		t.Fatalf("unmatched expenses did not preserve match isolation: %+v", expenses)
	}
}

func containsBankFeed(items []domain.UnmatchedBankTransaction, id int64) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func containsUnmatchedInvoice(items []domain.UnmatchedInvoice, id int64) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

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
