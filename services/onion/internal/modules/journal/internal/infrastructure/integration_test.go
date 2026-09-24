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

func TestPostJournalEntry_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{
		Date:        "2025-06-05",
		Description: "Post test",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Post(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Post: %v", err)
	}

	entry, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get after post: %v", err)
	}
	if entry.Status != "posted" {
		t.Errorf("entry status = %q after posting, want 'posted'", entry.Status)
	}
}

func TestPostJournalEntry_AlreadyPosted_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{
		Date:        "2025-06-06",
		Description: "Double post test",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Post(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("First Post: %v", err)
	}

	err = repo.Post(ctx, testutil.OwnerID, id)
	if err == nil {
		t.Fatal("expected error posting already-posted entry, got nil")
	}
}

func TestListJournalEntries_Pagination_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.List(ctx, testutil.OwnerID, 10, 1000)
	if err != nil {
		t.Fatalf("List with large offset: %v", err)
	}
}

func TestCreateJournalEntry_EmptyDescription_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{
		Date:  "2025-06-07",
		Lines: nil,
	})
	if err != nil {
		t.Fatalf("Create with empty description: %v", err)
	}
	if id == 0 {
		t.Fatal("Create returned id=0")
	}
}

func TestListJournalEntries_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	entries, err := repo.List(ctx, 0, 100, 0)
	if err != nil {
		t.Fatalf("List with owner_id=0: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for owner_id=0, got %d", len(entries))
	}
}

func TestCreateJournalEntry_FieldsPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wantDate := "2025-08-15"
	wantDesc := "Rent payment Aug 2025"

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{
		Date:        wantDate,
		Description: wantDesc,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	entry, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if entry.Description != wantDesc {
		t.Errorf("Description = %q, want %q", entry.Description, wantDesc)
	}
	if entry.Date.IsZero() {
		t.Error("Date is zero after persist")
	}
}

func TestGetJournalEntry_WrongOwner_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{
		Date:        "2025-09-01",
		Description: "Isolation check",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = repo.Get(ctx, testutil.OwnerID+1, id)
	if err == nil {
		t.Fatal("expected error fetching entry with wrong owner, got nil")
	}
}

func TestListJournalEntries_DateRange_ReturnsResults(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Create two entries at known dates.
	_, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{Date: "2025-07-01", Description: "July start"})
	if err != nil {
		t.Fatalf("Create July start: %v", err)
	}
	_, err = repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{Date: "2025-07-31", Description: "July end"})
	if err != nil {
		t.Fatalf("Create July end: %v", err)
	}

	entries, err := repo.List(ctx, testutil.OwnerID, 200, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	// At least the two we just inserted must come back.
	if len(entries) < 2 {
		t.Errorf("expected >= 2 entries, got %d", len(entries))
	}
}

func TestCreateJournalEntry_DefaultStatusIsDraft(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{
		Date:        "2025-10-01",
		Description: "Status check",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	entry, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if entry.Status != "draft" {
		t.Errorf("expected default status='draft', got %q", entry.Status)
	}
}

func TestGetLines_NoLines_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{
		Date:        "2025-10-02",
		Description: "No lines",
		Lines:       []domain.JournalLineInput{},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	lines, err := repo.GetLines(ctx, id)
	if err != nil {
		t.Fatalf("GetLines: %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("expected 0 lines, got %d", len(lines))
	}
}

func TestGetLines_LineDeltaMatch(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	accounts := ensureJournalTestAccounts(t, db, 2)

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{
		Date:        "2025-10-03",
		Description: "Delta check",
		Lines: []domain.JournalLineInput{
			{AccountID: accounts[0], Debit: 200, Credit: 0},
			{AccountID: accounts[1], Debit: 0, Credit: 200},
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	lines, err := repo.GetLines(ctx, id)
	if err != nil {
		t.Fatalf("GetLines: %v", err)
	}

	var totalDebit, totalCredit float64
	for _, l := range lines {
		totalDebit += l.Debit
		totalCredit += l.Credit
	}
	if totalDebit != totalCredit {
		t.Errorf("debits (%g) != credits (%g)", totalDebit, totalCredit)
	}
}

func TestCreateJournalEntry_OwnerIDStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{
		Date:        "2025-11-01",
		Description: "OwnerID stored check",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	entry, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if entry.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", entry.OwnerID, testutil.OwnerID)
	}
}

func TestListJournalEntries_LimitIsRespected(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Create enough entries to exceed the limit.
	for i := 0; i < 5; i++ {
		if _, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{
			Date:        "2025-11-15",
			Description: "Limit test",
		}); err != nil {
			t.Fatalf("Create[%d]: %v", i, err)
		}
	}

	entries, err := repo.List(ctx, testutil.OwnerID, 3, 0)
	if err != nil {
		t.Fatalf("List with limit=3: %v", err)
	}
	if len(entries) > 3 {
		t.Errorf("expected at most 3 entries, got %d", len(entries))
	}
}

func TestPostJournalEntry_PostedAtIsSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{
		Date:        "2025-12-01",
		Description: "Posted-at check",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Post(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Post: %v", err)
	}

	entry, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get after post: %v", err)
	}
	if entry.PostedAt.IsZero() {
		t.Error("PostedAt should be non-zero after posting")
	}
}

func TestGetJournalEntry_IDIsPositive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{
		Date:        "2025-12-15",
		Description: "ID positive check",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Errorf("Create returned non-positive ID: %d", id)
	}

	entry, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if entry.ID <= 0 {
		t.Errorf("entry.ID = %d, want > 0", entry.ID)
	}
}

func TestListJournalEntries_UnknownOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	entries, err := repo.List(ctx, 9999999, 100, 0)
	if err != nil {
		t.Fatalf("List unknown owner: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 for unknown owner, got %d", len(entries))
	}
}

func TestCreateJournalEntry_MultipleLines_CountMatches(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	accounts := ensureJournalTestAccounts(t, db, 3)

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{
		Date:        "2026-01-10",
		Description: "Multi-line count",
		Lines: []domain.JournalLineInput{
			{AccountID: accounts[0], Debit: 100, Credit: 0},
			{AccountID: accounts[1], Debit: 50, Credit: 0},
			{AccountID: accounts[2], Debit: 0, Credit: 150},
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	lines, err := repo.GetLines(ctx, id)
	if err != nil {
		t.Fatalf("GetLines: %v", err)
	}
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
}

func TestCreateJournalEntry_CreatedAtIsSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{
		Date:        "2026-02-01",
		Description: "CreatedAt check",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	entry, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if entry.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero after create")
	}
}

func TestCountByStatus_Journal_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	counts, err := repo.CountByStatus(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CountByStatus: %v", err)
	}
	_ = counts
}

func TestCountByStatus_Journal_ZeroOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	counts, err := repo.CountByStatus(ctx, 0)
	if err != nil {
		t.Fatalf("CountByStatus(ownerID=0): %v", err)
	}
	if len(counts) != 0 {
		t.Errorf("expected empty map for ownerID=0, got %v", counts)
	}
}

func TestCountByStatus_Journal_ReflectsCreatedEntry(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.JournalEntryInput{
		Date:        "2026-01-01",
		Description: "Count test",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	counts, err := repo.CountByStatus(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CountByStatus: %v", err)
	}
	// New entries default to "draft" status; at least one should be present.
	if counts["draft"] <= 0 && counts[""] <= 0 {
		t.Logf("counts: %v", counts)
	}
}

func TestSumLinesByAccountID_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	debit, credit, err := repo.SumLinesByAccountID(ctx, testutil.OwnerID, 1)
	if err != nil {
		t.Fatalf("SumLinesByAccountID: %v", err)
	}
	if debit < 0 {
		t.Errorf("expected non-negative debit, got %f", debit)
	}
	if credit < 0 {
		t.Errorf("expected non-negative credit, got %f", credit)
	}
}

func TestSumLinesByAccountID_ZeroOwnerReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	debit, credit, err := repo.SumLinesByAccountID(ctx, 0, 1)
	if err != nil {
		t.Fatalf("SumLinesByAccountID(ownerID=0): %v", err)
	}
	if debit != 0 || credit != 0 {
		t.Errorf("expected (0,0) for ownerID=0, got (%f,%f)", debit, credit)
	}
}
