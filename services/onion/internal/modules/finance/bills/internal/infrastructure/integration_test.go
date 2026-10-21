package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/voxire/onion/internal/modules/erp/purchase_orders"
	"testing"

	"github.com/voxire/onion/internal/modules/finance/bills/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/bills/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/testutil"
)

func TestList_Bills_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredBillRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestList_Bills_WrongOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredBillRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, 999999)
	if err != nil {
		t.Fatalf("List wrong owner: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items for unknown owner, got %d", len(items))
	}
}

func TestCreate_Bill_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredBillRepository(db)
	ctx := context.Background()

	input := domain.BillInput{
		Amount:   money.Amount(25000),
		Currency: "USD",
		DueDate:  "2026-12-31",
		Status:   "pending",
	}

	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreate_Bill_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredBillRepository(db)
	ctx := context.Background()

	input := domain.BillInput{
		Amount:   money.Amount(10000),
		Currency: "USD",
		DueDate:  "2026-11-30",
		Status:   "pending",
	}

	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	found := false
	for _, b := range items {
		if b.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created bill id=%d not found in list", id)
	}
}

func TestGet_Bill_ReturnsBill(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredBillRepository(db)
	ctx := context.Background()

	input := domain.BillInput{
		Amount:   money.Amount(7550),
		Currency: "USD",
		DueDate:  "2026-10-15",
		Status:   "pending",
	}

	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	bill, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if bill.ID != id {
		t.Errorf("ID = %d, want %d", bill.ID, id)
	}
}

func TestGet_Bill_WrongOwnerReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredBillRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BillInput{
		Amount: money.Amount(5000), Currency: "USD", DueDate: "2026-09-30", Status: "pending",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = repo.Get(ctx, 999999, id)
	if !errors.Is(err, domain.ErrBillNotFound) {
		t.Errorf("expected ErrBillNotFound, got %v", err)
	}
}

func TestGet_Bill_UnknownIDReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredBillRepository(db)
	ctx := context.Background()

	_, err := repo.Get(ctx, testutil.OwnerID, 999999999)
	if !errors.Is(err, domain.ErrBillNotFound) {
		t.Errorf("expected ErrBillNotFound, got %v", err)
	}
}

func TestUpdate_Bill_ChangesAmount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredBillRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BillInput{
		Amount: money.Amount(10000), Currency: "USD", DueDate: "2026-08-31", Status: "pending",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(ctx, testutil.OwnerID, id, domain.BillInput{
		Amount: money.Amount(20000), Currency: "USD", DueDate: "2026-08-31", Status: "pending",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	bill, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get after update: %v", err)
	}
	if bill.Amount != money.Amount(20000) {
		t.Errorf("Amount = %s, want 200.00", bill.Amount.String())
	}
}

func TestDelete_Bill_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredBillRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BillInput{
		Amount: money.Amount(30000), Currency: "USD", DueDate: "2026-07-31", Status: "pending",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestDelete_Bill_ThenNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredBillRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BillInput{
		Amount: money.Amount(15000), Currency: "USD", DueDate: "2026-06-30", Status: "pending",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = repo.Get(ctx, testutil.OwnerID, id)
	if !errors.Is(err, domain.ErrBillNotFound) {
		t.Errorf("expected ErrBillNotFound after delete, got %v", err)
	}
}

func configuredBillRepository(db *sql.DB) *infrastructure.Repository {
	return infrastructure.NewRepository(db).WithPurchaseOrderReader(purchase_orders.NewModule(context.Background(), db).Exists)
}
