package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/finance_reconciliation/internal/domain"
	"github.com/voxire/onion/internal/modules/finance_reconciliation/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

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

func TestCountMatches_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cnt, err := repo.CountMatches(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CountMatches: %v", err)
	}
	if cnt < 0 {
		t.Errorf("expected non-negative count, got %d", cnt)
	}
}

func TestCountMatches_ZeroOwnerReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cnt, err := repo.CountMatches(ctx, 0)
	if err != nil {
		t.Fatalf("CountMatches(ownerID=0): %v", err)
	}
	if cnt != 0 {
		t.Errorf("expected 0 for ownerID=0, got %d", cnt)
	}
}

func TestCountMatchesByTargetType_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	counts, err := repo.CountMatchesByTargetType(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CountMatchesByTargetType: %v", err)
	}
	_ = counts
}

func TestCountMatchesByTargetType_ZeroOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	counts, err := repo.CountMatchesByTargetType(ctx, 0)
	if err != nil {
		t.Fatalf("CountMatchesByTargetType(ownerID=0): %v", err)
	}
	if len(counts) != 0 {
		t.Errorf("expected empty map for ownerID=0, got %v", counts)
	}
}

func TestCreateMatchUpdatesBankFeedProjection(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	businessID := time.Now().UnixNano()
	externalID := fmt.Sprintf("projection-feed-%d", businessID)

	insert := func(query string, args ...any) int64 {
		result, err := db.ExecContext(ctx, query, args...)
		if err != nil {
			t.Fatalf("insert fixture: %v", err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("fixture id: %v", err)
		}
		return id
	}

	bankFeedID := insert(
		"INSERT INTO finance_bank_feeds (owner_id, business_id, bank_account_id, amount_minor_units, occurred_at, external_id) VALUES (?, ?, ?, ?, NOW(), ?)",
		businessID, businessID, 1, 12345, externalID,
	)
	invoiceID := insert(
		"INSERT INTO finance_invoices (owner_id, business_id, invoice_number, client_name, issue_date, due_date, total) VALUES (?, ?, ?, ?, CURDATE(), CURDATE(), ?)",
		businessID, businessID, fmt.Sprintf("PROJECTION-%d", businessID), "Projection client", 123.45,
	)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM finance_reconciliation_matches WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM finance_reconciliation_targets WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM finance_bank_feeds WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM finance_invoices WHERE business_id = ?", businessID)
	})

	if _, err := repo.CreateMatch(ctx, businessID, domain.MatchInput{
		BankFeedID: bankFeedID,
		TargetType: domain.TargetTypeInvoice,
		TargetID:   invoiceID,
	}); err != nil {
		t.Fatalf("CreateMatch: %v", err)
	}

	var matchedInvoiceID, matchedExpenseID sql.NullInt64
	if err := db.QueryRowContext(ctx,
		"SELECT matched_invoice_id, matched_expense_id FROM finance_bank_feeds WHERE id = ? AND business_id = ?",
		bankFeedID, businessID,
	).Scan(&matchedInvoiceID, &matchedExpenseID); err != nil {
		t.Fatalf("read bank-feed projection: %v", err)
	}
	if !matchedInvoiceID.Valid || matchedInvoiceID.Int64 != invoiceID {
		t.Fatalf("matched_invoice_id = %+v, want %d", matchedInvoiceID, invoiceID)
	}
	if matchedExpenseID.Valid {
		t.Fatalf("matched_expense_id = %+v, want NULL", matchedExpenseID)
	}

	feeds, err := repo.ListUnmatchedBankFeeds(ctx, businessID)
	if err != nil {
		t.Fatalf("ListUnmatchedBankFeeds after match: %v", err)
	}
	if containsBankFeed(feeds, bankFeedID) {
		t.Fatal("matched bank feed remained in the unmatched projection")
	}

	if _, err := repo.CreateMatch(ctx, businessID, domain.MatchInput{
		BankFeedID: bankFeedID,
		TargetType: domain.TargetTypeInvoice,
		TargetID:   invoiceID,
	}); !errors.Is(err, domain.ErrAlreadyMatched) {
		t.Fatalf("second match error = %v, want ErrAlreadyMatched", err)
	}
}
