package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/store_checkout/internal/domain"
	"github.com/voxire/onion/internal/modules/store_checkout/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCheckout_Create_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	input := domain.CheckoutCreateInput{
		SessionID:     "sess-abc-123",
		CustomerEmail: "test@example.com",
		Items:         `[{"product_id":1,"qty":2}]`,
		Subtotal:      50.0,
		Tax:           5.0,
		Total:         55.0,
	}
	id, err := repo.Create(context.Background(), testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCheckout_Create_EmptyItems(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	input := domain.CheckoutCreateInput{
		Subtotal: 10.0,
		Tax:      1.0,
		Total:    11.0,
	}
	id, err := repo.Create(context.Background(), testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create with empty items: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCheckout_Create_MultipleCheckouts(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	for i := 0; i < 3; i++ {
		id, err := repo.Create(context.Background(), testutil.OwnerID, domain.CheckoutCreateInput{
			Subtotal: float64(i) * 10,
			Tax:      float64(i),
			Total:    float64(i)*10 + float64(i),
		})
		if err != nil {
			t.Fatalf("Create #%d: %v", i, err)
		}
		if id <= 0 {
			t.Errorf("iteration %d: expected positive ID", i)
		}
	}
}

func TestCheckout_Complete_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.CheckoutCreateInput{
		Subtotal: 20.0,
		Tax:      2.0,
		Total:    22.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Complete(context.Background(), testutil.OwnerID, id); err != nil {
		t.Fatalf("Complete: %v", err)
	}
}

func TestCheckout_Complete_UnknownID_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	// Completing a non-existent checkout should not return an error (0 rows affected)
	err := repo.Complete(context.Background(), testutil.OwnerID, 9999999)
	if err != nil {
		t.Fatalf("Complete on missing ID: %v", err)
	}
}

func TestCheckout_Create_WithSessionID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.CheckoutCreateInput{
		SessionID: "unique-session-xyz",
		Subtotal:  100.0,
		Tax:       10.0,
		Total:     110.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestCheckout_Create_ZeroTotals(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.CheckoutCreateInput{
		Subtotal: 0,
		Tax:      0,
		Total:    0,
	})
	if err != nil {
		t.Fatalf("Create with zero totals: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestCheckout_Complete_CrossOwner_NoEffect(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.CheckoutCreateInput{
		Subtotal: 30.0,
		Tax:      3.0,
		Total:    33.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	// Complete with wrong owner should silently do nothing
	err = repo.Complete(context.Background(), 9999999, id)
	if err != nil {
		t.Fatalf("Complete cross-owner: %v", err)
	}
}
