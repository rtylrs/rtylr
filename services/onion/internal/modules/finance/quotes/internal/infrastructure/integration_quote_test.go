package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/finance/quotes/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/quotes/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetQuote_NotFound_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.Get(ctx, testutil.OwnerID, 9999999)
	if err == nil {
		t.Fatal("expected error for non-existent quote id=9999999, got nil")
	}
}

func TestUpdateQuote_ChangesClientName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.QuoteInput{
		CustomerName: "Old Client Name",
		Status:       "draft",
	})
	if err != nil {
		t.Fatalf("quotes.Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	if err := repo.Update(ctx, testutil.OwnerID, id, domain.QuoteInput{
		CustomerName: "New Client Name",
		Status:       "draft",
	}); err != nil {
		t.Fatalf("quotes.Update: %v", err)
	}

	q, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("quotes.Get after update: %v", err)
	}
	if q.CustomerName != "New Client Name" {
		t.Errorf("expected CustomerName=New Client Name, got %q", q.CustomerName)
	}
}

func TestDeleteQuote_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.QuoteInput{
		CustomerName: "Delete Row Customer",
		Status:       "draft",
	})
	if err != nil {
		t.Fatalf("quotes.Create: %v", err)
	}

	if err := repo.Delete(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("quotes.Delete: %v", err)
	}

	_, err = repo.Get(ctx, testutil.OwnerID, id)
	if err == nil {
		t.Fatal("expected error after deletion, got nil")
	}
}

func TestDeleteQuote_NonExistentReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Deleting a non-existent quote should report the same result as other
	// quote-by-id operations rather than silently claiming success.
	err := repo.Delete(ctx, testutil.OwnerID, 9999999)
	if !errors.Is(err, domain.ErrQuoteNotFound) {
		t.Fatalf("quotes.Delete(non-existent): expected ErrQuoteNotFound, got %v", err)
	}
}

func TestCreate_ThenGet_Roundtrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.QuoteInput{
		CustomerName:  "Roundtrip Customer",
		CustomerEmail: "roundtrip@example.com",
		Status:        "draft",
		Total:         250.0,
	}
	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("quotes.Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	q, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("quotes.Get: %v", err)
	}
	if q.CustomerName != input.CustomerName {
		t.Errorf("CustomerName: got %q, want %q", q.CustomerName, input.CustomerName)
	}
	if q.CustomerEmail != input.CustomerEmail {
		t.Errorf("CustomerEmail: got %q, want %q", q.CustomerEmail, input.CustomerEmail)
	}
	if q.Total != input.Total {
		t.Errorf("Total: got %.2f, want %.2f", q.Total, input.Total)
	}
}

func TestListQuotes_Pagination(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Seed at least 4 quotes so pagination is testable
	for i := 0; i < 4; i++ {
		_, err := repo.Create(ctx, testutil.OwnerID, domain.QuoteInput{
			CustomerName: "Pagination Test Customer",
			Status:       "draft",
		})
		if err != nil {
			t.Fatalf("quotes.Create[%d]: %v", i, err)
		}
	}

	page0, err := repo.List(ctx, testutil.OwnerID, 2, 0)
	if err != nil {
		t.Fatalf("List(limit=2, offset=0): %v", err)
	}
	page1, err := repo.List(ctx, testutil.OwnerID, 2, 2)
	if err != nil {
		t.Fatalf("List(limit=2, offset=2): %v", err)
	}

	if len(page0) == 0 || len(page1) == 0 {
		t.Skip("not enough rows to test pagination")
	}
	if page0[0].ID == page1[0].ID {
		t.Error("pagination offset not working: first items on page 0 and page 1 are identical")
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

func TestCountByStatus_Quotes_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	counts, err := repo.CountByStatus(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CountByStatus: %v", err)
	}
	_ = counts
}

func TestCountByStatus_Quotes_ZeroOwnerReturnsEmpty(t *testing.T) {
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

func TestCountByStatus_Quotes_ReflectsCreatedQuote(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.QuoteInput{
		CustomerName: "CountByStatus Customer",
		Status:       "draft",
		Total:        100,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	counts, err := repo.CountByStatus(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CountByStatus: %v", err)
	}
	if counts["draft"] <= 0 {
		t.Errorf("expected at least 1 draft quote, got %d", counts["draft"])
	}
}
