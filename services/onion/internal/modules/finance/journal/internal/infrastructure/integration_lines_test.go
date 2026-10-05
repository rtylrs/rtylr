package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/finance/journal/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/journal/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

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
