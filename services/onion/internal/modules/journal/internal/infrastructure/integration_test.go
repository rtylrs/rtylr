package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/journal/internal/domain"
	"github.com/voxire/onion/internal/modules/journal/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func ensureJournalTestAccounts(t *testing.T, db *sql.DB, count int) []int64 {
	t.Helper()
	ids := make([]int64, 0, count)
	for i := 0; i < count; i++ {
		result, err := db.ExecContext(context.Background(), `
			INSERT INTO finance_accounts (owner_id, business_id, name, type, currency, balance, is_active)
			VALUES (?, ?, ?, 'asset', 'USD', 0, TRUE)`,
			testutil.OwnerID, testutil.BusinessID, fmt.Sprintf("Journal test account %d", i+1))
		if err != nil {
			t.Fatalf("insert journal test account: %v", err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("journal test account id: %v", err)
		}
		ids = append(ids, id)
		t.Cleanup(func() {
			_, _ = db.ExecContext(context.Background(), `DELETE FROM finance_accounts WHERE id = ?`, id)
		})
	}
	return ids
}

func TestListJournalEntries_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	entries, err := repo.List(ctx, testutil.OwnerID, 50, 0)
	if err != nil {
		t.Fatalf("List journal entries: %v", err)
	}
	_ = entries
}

func TestListJournalEntries_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	entries, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("List journal entries: %v", err)
	}
	for _, e := range entries {
		if e.OwnerID != testutil.OwnerID {
			t.Errorf("entry %d has owner_id %d, want %d", e.ID, e.OwnerID, testutil.OwnerID)
		}
	}
}

func TestCreateJournalEntry_NoLines_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{
		Date:        "2025-06-01",
		Description: "Test entry no lines",
		Lines:       []domain.JournalLineInput{},
	})
	if err != nil {
		t.Fatalf("Create journal entry: %v", err)
	}
	if id == 0 {
		t.Fatal("Create returned id=0")
	}
}

func TestCreateJournalEntry_WithLines_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	accounts := ensureJournalTestAccounts(t, db, 2)

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{
		Date:        "2025-06-02",
		Description: "Test entry with lines",
		Lines: []domain.JournalLineInput{
			{AccountID: accounts[0], Debit: 100.00, Credit: 0},
			{AccountID: accounts[1], Debit: 0, Credit: 100.00},
		},
	})
	if err != nil {
		t.Fatalf("Create journal entry with lines: %v", err)
	}
	if id == 0 {
		t.Fatal("Create returned id=0")
	}
}

func TestCreateJournalEntry_RejectsAccountFromAnotherBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const (
		businessA int64 = 907301
		businessB int64 = 907302
		ownerA    int64 = 907201
		ownerB    int64 = 907202
	)
	testutil.EnsureBusiness(t, db, businessA, ownerA)
	testutil.EnsureBusiness(t, db, businessB, ownerB)

	insertAccount := func(businessID, ownerID int64, name string) int64 {
		t.Helper()
		result, err := db.ExecContext(ctx, `
			INSERT INTO finance_accounts (owner_id, business_id, name, type, currency, balance, is_active)
			VALUES (?, ?, ?, 'asset', 'USD', 0, TRUE)`, ownerID, businessID, name)
		if err != nil {
			t.Fatalf("insert account for business %d: %v", businessID, err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("account id for business %d: %v", businessID, err)
		}
		t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), `DELETE FROM finance_accounts WHERE id = ?`, id) })
		return id
	}
	accountA := insertAccount(businessA, ownerA, "Business A cash")
	accountB := insertAccount(businessB, ownerB, "Business B cash")

	_, err := repo.Create(ctx, ownerA, businessA, domain.JournalEntryInput{
		Date: "2026-09-20",
		Lines: []domain.JournalLineInput{
			{AccountID: accountA, Debit: 100},
			{AccountID: accountB, Credit: 100},
		},
	})
	if !errors.Is(err, domain.ErrInvalidJournalEntry) {
		t.Fatalf("cross-business account error = %v, want ErrInvalidJournalEntry", err)
	}

	var entries int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM finance_journal_entries WHERE business_id = ?`, businessA).Scan(&entries); err != nil {
		t.Fatalf("count rejected journal entries: %v", err)
	}
	if entries != 0 {
		t.Fatalf("rejected cross-tenant entry left %d header rows", entries)
	}
}

func TestGetJournalEntry_AfterCreate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{
		Date:        "2025-06-03",
		Description: "Get after create",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	entry, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if entry.ID != id {
		t.Errorf("Get returned id %d, want %d", entry.ID, id)
	}
	if entry.OwnerID != testutil.OwnerID {
		t.Errorf("Get returned owner_id %d, want %d", entry.OwnerID, testutil.OwnerID)
	}
}

func TestGetJournalEntry_NotFound_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.Get(ctx, testutil.OwnerID, 9999999)
	if err == nil {
		t.Fatal("expected error for non-existent journal entry, got nil")
	}
}

func TestGetLines_AfterCreateWithLines(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	accounts := ensureJournalTestAccounts(t, db, 2)

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{
		Date:        "2025-06-04",
		Description: "Lines test",
		Lines: []domain.JournalLineInput{
			{AccountID: accounts[0], Debit: 50.00, Credit: 0},
			{AccountID: accounts[1], Debit: 0, Credit: 50.00},
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	lines, err := repo.GetLines(ctx, id)
	if err != nil {
		t.Fatalf("GetLines: %v", err)
	}
	if len(lines) != 2 {
		t.Errorf("GetLines returned %d lines, want 2", len(lines))
	}
}
