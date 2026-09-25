package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/bills/internal/domain"
	"github.com/voxire/onion/internal/modules/bills/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/testutil"
)

func TestUpdate_Bill_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BillInput{
		Amount:  money.Amount(5500),
		DueDate: "2026-09-01",
		Status:  "pending",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(ctx, testutil.OwnerID, id, domain.BillInput{
		Amount:  money.Amount(5500),
		DueDate: "2026-09-01",
		Status:  "paid",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	bill, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get after update: %v", err)
	}
	if bill.Status != "paid" {
		t.Errorf("Status = %q, want %q", bill.Status, "paid")
	}
}

func TestUpdate_Bill_ChangesDescription(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BillInput{
		Amount:      money.Amount(6000),
		Currency:    "USD",
		DueDate:     "2026-08-15",
		Status:      "pending",
		Description: "Original description",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(ctx, testutil.OwnerID, id, domain.BillInput{
		Amount:      money.Amount(6000),
		Currency:    "USD",
		DueDate:     "2026-08-15",
		Status:      "pending",
		Description: "Updated description",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	bill, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get after update: %v", err)
	}
	if bill.Description != "Updated description" {
		t.Errorf("Description = %q, want %q", bill.Description, "Updated description")
	}
}

func TestCreate_Bill_DefaultStatusPending(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Omit status; repository should default to "pending".
	id, err := repo.Create(ctx, testutil.OwnerID, domain.BillInput{
		Amount:  money.Amount(3500),
		DueDate: "2026-07-01",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	bill, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if bill.Status != "pending" {
		t.Errorf("Status = %q, want %q", bill.Status, "pending")
	}
}

func TestList_Bills_ItemsHaveNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, testutil.OwnerID, domain.BillInput{
		Amount: money.Amount(1100), DueDate: "2026-06-01", Status: "pending",
	})

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, b := range items {
		if b.ID == 0 {
			t.Errorf("items[%d].ID is 0", i)
		}
	}
}

func TestCreate_Bill_CurrencyEUR(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BillInput{
		Amount:   money.Amount(8000),
		Currency: "EUR",
		DueDate:  "2026-05-31",
		Status:   "pending",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	bill, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if bill.Currency != "EUR" {
		t.Errorf("Currency = %q, want %q", bill.Currency, "EUR")
	}
}

func TestCreate_ThenList_FindsNewBill(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	beforeItems, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}
	beforeCount := len(beforeItems)

	_, err = repo.Create(ctx, testutil.OwnerID, domain.BillInput{
		Amount:  money.Amount(99900),
		DueDate: "2026-12-31",
		Status:  "pending",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	afterItems, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(afterItems) <= beforeCount {
		t.Errorf("expected more items after create; before=%d after=%d", beforeCount, len(afterItems))
	}
}

func TestSumByStatus_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	sum, err := repo.SumByStatus(ctx, testutil.OwnerID, "pending")
	if err != nil {
		t.Fatalf("SumByStatus: %v", err)
	}
	if sum < money.Amount(0) {
		t.Errorf("expected non-negative sum, got %s", sum.String())
	}
}

func TestSumByStatus_ZeroOwnerReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	sum, err := repo.SumByStatus(ctx, 0, "pending")
	if err != nil {
		t.Fatalf("SumByStatus(ownerID=0): %v", err)
	}
	if sum != money.Amount(0) {
		t.Errorf("expected 0 for ownerID=0, got %s", sum.String())
	}
}

func TestSumByStatus_ReflectsCreatedBill(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BillInput{
		Amount:  money.Amount(12345),
		DueDate: "2026-12-31",
		Status:  "pending",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	sum, err := repo.SumByStatus(ctx, testutil.OwnerID, "pending")
	if err != nil {
		t.Fatalf("SumByStatus: %v", err)
	}
	if sum < money.Amount(12345) {
		t.Errorf("expected sum >= 123.45, got %s", sum.String())
	}
}

func TestCountByStatus_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	counts, err := repo.CountByStatus(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CountByStatus: %v", err)
	}
	_ = counts
}

func TestCountByStatus_ZeroOwnerReturnsEmpty(t *testing.T) {
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

func TestCountByStatus_ReflectsCreatedBill(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BillInput{
		Amount:  money.Amount(5000),
		DueDate: "2026-11-30",
		Status:  "pending",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	counts, err := repo.CountByStatus(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CountByStatus: %v", err)
	}
	if counts["pending"] <= 0 {
		t.Errorf("expected at least 1 pending bill, got %d", counts["pending"])
	}
}
