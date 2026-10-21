package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/finance/quotes/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/quotes/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestUpdateStatus_NonExistentQuoteReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// The UPDATE matches zero rows, which must not be reported as success.
	err := repo.UpdateStatus(ctx, testutil.OwnerID, 9999999, "sent")
	if !errors.Is(err, domain.ErrQuoteNotFound) {
		t.Fatalf("quotes.UpdateStatus(non-existent): expected ErrQuoteNotFound, got %v", err)
	}
}

func TestUpdateStatus_OtherBusinessQuoteReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.QuoteInput{
		CustomerName: "Other Business Quote",
		Status:       "draft",
	})
	if err != nil {
		t.Fatalf("quotes.Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	err = repo.UpdateStatus(ctx, testutil.OwnerID+9999, id, "sent")
	if !errors.Is(err, domain.ErrQuoteNotFound) {
		t.Fatalf("quotes.UpdateStatus(foreign business): expected ErrQuoteNotFound, got %v", err)
	}
}

func TestUpdateStatus_UpdatesExistingQuote(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.QuoteInput{
		CustomerName: "Status Update Customer",
		Status:       "draft",
	})
	if err != nil {
		t.Fatalf("quotes.Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	if err = repo.UpdateStatus(ctx, testutil.OwnerID, id, "sent"); err != nil {
		t.Fatalf("quotes.UpdateStatus: %v", err)
	}

	q, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("quotes.Get: %v", err)
	}
	if q.Status != "sent" {
		t.Errorf("expected status=sent, got %q", q.Status)
	}
}

func TestUpdateStatus_SameStatusIsIdempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.QuoteInput{
		CustomerName: "Idempotent Status Customer",
		Status:       "draft",
	})
	if err != nil {
		t.Fatalf("quotes.Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	// MySQL reports changed rows, not matched rows: re-setting the current
	// status affects zero rows but the quote does exist.
	if err = repo.UpdateStatus(ctx, testutil.OwnerID, id, "draft"); err != nil {
		t.Fatalf("quotes.UpdateStatus(same status): expected no error, got %v", err)
	}

	q, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("quotes.Get: %v", err)
	}
	if q.Status != "draft" {
		t.Errorf("expected status=draft, got %q", q.Status)
	}
}
