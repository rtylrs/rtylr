package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/finance/quotes/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/quotes/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

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

func TestListQuotes_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	quotes, err := repo.List(ctx, 0, 20, 0)
	if err != nil {
		t.Fatalf("quotes.List(ownerID=0): %v", err)
	}
	if len(quotes) != 0 {
		t.Fatalf("expected 0 quotes for ownerID=0, got %d", len(quotes))
	}
}

func TestListQuotes_IsNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	quotes, err := repo.List(ctx, testutil.OwnerID, 20, 0)
	if err != nil {
		t.Fatalf("quotes.List: %v", err)
	}
	if quotes == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
}

func TestCreateQuote_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.QuoteInput{
		CustomerName: "Persist Test",
		Status:       "draft",
		Total:        75.0,
	})
	if err != nil {
		t.Fatalf("quotes.Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })
}

func TestCreateQuote_ClientNameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const wantName = "ClientNameStore Customer"
	id, err := repo.Create(ctx, testutil.OwnerID, domain.QuoteInput{
		CustomerName: wantName,
		Status:       "draft",
	})
	if err != nil {
		t.Fatalf("quotes.Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	q, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("quotes.Get: %v", err)
	}
	if q.CustomerName != wantName {
		t.Errorf("expected CustomerName=%q, got %q", wantName, q.CustomerName)
	}
}

func TestCreateQuote_TotalAmountIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const wantTotal = 123.45
	id, err := repo.Create(ctx, testutil.OwnerID, domain.QuoteInput{
		CustomerName: "Total Store Customer",
		Status:       "draft",
		Total:        wantTotal,
	})
	if err != nil {
		t.Fatalf("quotes.Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	q, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("quotes.Get: %v", err)
	}
	if q.Total != wantTotal {
		t.Errorf("expected Total=%.2f, got %.2f", wantTotal, q.Total)
	}
}

func TestGetQuote_ByID_ReturnsCorrectData(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.QuoteInput{
		CustomerName: "ByID Customer",
		Status:       "draft",
		Total:        50.0,
	})
	if err != nil {
		t.Fatalf("quotes.Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	q, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("quotes.Get: %v", err)
	}
	if q.ID != id {
		t.Errorf("expected ID=%d, got %d", id, q.ID)
	}
	if q.CustomerName != "ByID Customer" {
		t.Errorf("expected CustomerName=ByID Customer, got %q", q.CustomerName)
	}
}
