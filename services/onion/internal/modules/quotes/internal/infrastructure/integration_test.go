package infrastructure_test

import (
	"context"
	"errors"
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

func TestUpdateQuote_ConvertedIsNotEditable(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.QuoteInput{
		CustomerName: "Converted Quote",
		Status:       "converted",
	})
	if err != nil {
		t.Fatalf("quotes.Create: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM finance_quotes WHERE id = ? AND business_id = ?", id, testutil.OwnerID)
	})

	err = repo.Update(ctx, testutil.OwnerID, id, domain.QuoteInput{
		CustomerName: "Should Not Change",
		Status:       "converted",
	})
	if !errors.Is(err, domain.ErrQuoteNotEditable) {
		t.Fatalf("quotes.Update converted: expected ErrQuoteNotEditable, got %v", err)
	}
}

func TestDeleteQuote_ConvertedIsNotEditable(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.QuoteInput{
		CustomerName: "Converted Quote",
		Status:       "converted",
	})
	if err != nil {
		t.Fatalf("quotes.Create: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM finance_quotes WHERE id = ? AND business_id = ?", id, testutil.OwnerID)
	})

	err = repo.Delete(ctx, testutil.OwnerID, id)
	if !errors.Is(err, domain.ErrQuoteNotEditable) {
		t.Fatalf("quotes.Delete converted: expected ErrQuoteNotEditable, got %v", err)
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
