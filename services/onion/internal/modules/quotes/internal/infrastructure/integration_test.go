package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/quotes/internal/domain"
	"github.com/voxire/onion/internal/modules/quotes/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListQuotes_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	quotes, err := repo.List(ctx, testutil.OwnerID, 20, 0)
	if err != nil {
		t.Fatalf("quotes.List: unexpected error: %v", err)
	}
	if quotes == nil {
		t.Fatal("quotes.List: expected non-nil slice, got nil")
	}
}

func TestListQuotes_WrongOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	quotes, err := repo.List(ctx, 0, 20, 0)
	if err != nil {
		t.Fatalf("quotes.List(ownerID=0): unexpected error: %v", err)
	}
	if len(quotes) != 0 {
		t.Fatalf("quotes.List(ownerID=0): expected 0 rows, got %d", len(quotes))
	}
}

func TestCreateQuote_ReturnsInsertedID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.QuoteInput{
		CustomerName: "Test Customer",
		Status:       "draft",
		Subtotal:     100.0,
		Total:        100.0,
	}

	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("quotes.Create: unexpected error: %v", err)
	}
	if id <= 0 {
		t.Fatalf("quotes.Create: expected positive ID, got %d", id)
	}
}

func TestCreateQuote_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.QuoteInput{
		CustomerName:  "List Check Customer",
		CustomerEmail: "listcheck@example.com",
		Status:        "draft",
		Total:         50.0,
	}

	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("quotes.Create: %v", err)
	}

	quotes, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("quotes.List: %v", err)
	}

	var found bool
	for _, q := range quotes {
		if q.ID == id {
			found = true
			if q.CustomerName != input.CustomerName {
				t.Errorf("CustomerName mismatch: got %q, want %q", q.CustomerName, input.CustomerName)
			}
			break
		}
	}
	if !found {
		t.Fatalf("created quote id=%d not found in list", id)
	}
}

func TestGetQuote_ExistingQuote(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.QuoteInput{
		CustomerName: "Get Test Customer",
		Status:       "draft",
		Total:        200.0,
	}

	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("quotes.Create: %v", err)
	}

	q, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("quotes.Get(id=%d): unexpected error: %v", id, err)
	}
	if q.ID != id {
		t.Errorf("ID mismatch: got %d, want %d", q.ID, id)
	}
	if q.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID mismatch: got %d, want %d", q.OwnerID, testutil.OwnerID)
	}
}

func TestGetQuote_NotFound_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.Get(ctx, testutil.OwnerID, 0)
	if err == nil {
		t.Fatal("quotes.Get(id=0): expected error, got nil")
	}
}

func TestUpdateQuote_ChangesFields(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.QuoteInput{
		CustomerName: "Before Update",
		Status:       "draft",
	})
	if err != nil {
		t.Fatalf("quotes.Create: %v", err)
	}

	err = repo.Update(ctx, testutil.OwnerID, id, domain.QuoteInput{
		CustomerName: "After Update",
		Status:       "sent",
		Total:        300.0,
	})
	if err != nil {
		t.Fatalf("quotes.Update: %v", err)
	}

	q, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("quotes.Get after update: %v", err)
	}
	if q.CustomerName != "After Update" {
		t.Errorf("CustomerName not updated: got %q", q.CustomerName)
	}
	if q.Status != "sent" {
		t.Errorf("Status not updated: got %q", q.Status)
	}
}

func TestDeleteQuote_RemovedFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.QuoteInput{
		CustomerName: "To Be Deleted",
		Status:       "draft",
	})
	if err != nil {
		t.Fatalf("quotes.Create: %v", err)
	}

	err = repo.Delete(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("quotes.Delete: %v", err)
	}

	_, err = repo.Get(ctx, testutil.OwnerID, id)
	if err == nil {
		t.Fatalf("quotes.Get after delete: expected error, got nil")
	}
}

func TestListQuotes_PaginationOffset(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Seed a few quotes
	for i := 0; i < 3; i++ {
		_, err := repo.Create(ctx, testutil.OwnerID, domain.QuoteInput{
			CustomerName: "Pagination Customer",
			Status:       "draft",
		})
		if err != nil {
			t.Fatalf("quotes.Create[%d]: %v", i, err)
		}
	}

	page0, err := repo.List(ctx, testutil.OwnerID, 2, 0)
	if err != nil {
		t.Fatalf("List page 0: %v", err)
	}
	page1, err := repo.List(ctx, testutil.OwnerID, 2, 2)
	if err != nil {
		t.Fatalf("List page 1: %v", err)
	}

	if len(page0) > 0 && len(page1) > 0 {
		if page0[0].ID == page1[0].ID {
			t.Error("pagination offset not working: first items on both pages are identical")
		}
	}
}

func TestCreateQuote_DefaultStatusIsDraft(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Status left empty -- repo should default to "draft"
	id, err := repo.Create(ctx, testutil.OwnerID, domain.QuoteInput{
		CustomerName: "Default Status Customer",
	})
	if err != nil {
		t.Fatalf("quotes.Create: %v", err)
	}

	q, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("quotes.Get: %v", err)
	}
	if q.Status != "draft" {
		t.Errorf("expected default status=draft, got %q", q.Status)
	}
}

func TestListQuotes_ScopedToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, testutil.OwnerID, domain.QuoteInput{
		CustomerName: "Scoped Quote",
		Status:       "draft",
	})
	if err != nil {
		t.Fatalf("quotes.Create: %v", err)
	}

	other, err := repo.List(ctx, testutil.OwnerID+9999, 50, 0)
	if err != nil {
		t.Fatalf("quotes.List(other owner): %v", err)
	}
	for _, q := range other {
		if q.OwnerID == testutil.OwnerID {
			t.Errorf("other-owner list contains ownerID=%d record", testutil.OwnerID)
		}
	}
}
