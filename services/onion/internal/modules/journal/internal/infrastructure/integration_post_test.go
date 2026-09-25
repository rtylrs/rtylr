package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/journal/internal/domain"
	"github.com/voxire/onion/internal/modules/journal/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

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
