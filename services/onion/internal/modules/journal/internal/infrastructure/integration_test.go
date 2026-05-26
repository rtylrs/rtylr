package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/journal/internal/domain"
	"github.com/voxire/onion/internal/modules/journal/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

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

	id, err := repo.Create(ctx, testutil.OwnerID, domain.JournalEntryInput{
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

	id, err := repo.Create(ctx, testutil.OwnerID, domain.JournalEntryInput{
		Date:        "2025-06-02",
		Description: "Test entry with lines",
		Lines: []domain.JournalLineInput{
			{AccountID: 1, Debit: 100.00, Credit: 0},
			{AccountID: 2, Debit: 0, Credit: 100.00},
		},
	})
	if err != nil {
		t.Fatalf("Create journal entry with lines: %v", err)
	}
	if id == 0 {
		t.Fatal("Create returned id=0")
	}
}

func TestGetJournalEntry_AfterCreate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.JournalEntryInput{
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

	id, err := repo.Create(ctx, testutil.OwnerID, domain.JournalEntryInput{
		Date:        "2025-06-04",
		Description: "Lines test",
		Lines: []domain.JournalLineInput{
			{AccountID: 1, Debit: 50.00, Credit: 0},
			{AccountID: 3, Debit: 0, Credit: 50.00},
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

	id, err := repo.Create(ctx, testutil.OwnerID, domain.JournalEntryInput{
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

	id, err := repo.Create(ctx, testutil.OwnerID, domain.JournalEntryInput{
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

	id, err := repo.Create(ctx, testutil.OwnerID, domain.JournalEntryInput{
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
