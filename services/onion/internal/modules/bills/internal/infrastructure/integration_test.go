package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/bills/internal/domain"
	"github.com/voxire/onion/internal/modules/bills/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/testutil"
)

func TestList_Bills_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
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
	repo := infrastructure.NewRepository(db)
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
	repo := infrastructure.NewRepository(db)
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
	repo := infrastructure.NewRepository(db)
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
	repo := infrastructure.NewRepository(db)
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
	repo := infrastructure.NewRepository(db)
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
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.Get(ctx, testutil.OwnerID, 999999999)
	if !errors.Is(err, domain.ErrBillNotFound) {
		t.Errorf("expected ErrBillNotFound, got %v", err)
	}
}

func TestUpdate_Bill_ChangesAmount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
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
	repo := infrastructure.NewRepository(db)
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
	repo := infrastructure.NewRepository(db)
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

func TestDelete_Bill_WrongOwnerReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BillInput{
		Amount: money.Amount(9000), Currency: "USD", DueDate: "2026-05-31", Status: "pending",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = repo.Delete(ctx, 999999, id)
	if !errors.Is(err, domain.ErrBillNotFound) {
		t.Errorf("expected ErrBillNotFound, got %v", err)
	}
}

func TestList_Bills_ItemsHaveOwnerID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, testutil.OwnerID, domain.BillInput{
		Amount: money.Amount(1000), Currency: "USD", DueDate: "2026-04-30", Status: "pending",
	})

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, b := range items {
		if b.OwnerID != testutil.OwnerID {
			t.Errorf("items[%d].OwnerID = %d, want %d", i, b.OwnerID, testutil.OwnerID)
		}
	}
}

func TestCreate_Bill_DefaultCurrencyUSD(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BillInput{
		Amount:  money.Amount(2000),
		DueDate: "2026-03-31",
		Status:  "pending",
		// Currency intentionally blank -> defaults to USD
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	bill, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if bill.Currency != "USD" {
		t.Errorf("Currency = %q, want %q", bill.Currency, "USD")
	}
}

func TestList_Bills_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("List(0): %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items for owner 0, got %d", len(items))
	}
}

func TestCreate_Bill_AmountStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BillInput{
		Amount:   money.Amount(45678),
		Currency: "USD",
		DueDate:  "2026-12-01",
		Status:   "pending",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	bill, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if bill.Amount != money.Amount(45678) {
		t.Errorf("Amount = %s, want 456.78", bill.Amount.String())
	}
}

func TestCreate_Bill_VendorStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	desc := "Electricity bill"
	id, err := repo.Create(ctx, testutil.OwnerID, domain.BillInput{
		Amount:      money.Amount(20000),
		Currency:    "USD",
		DueDate:     "2026-11-15",
		Status:      "pending",
		Description: desc,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	bill, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if bill.Description != desc {
		t.Errorf("Description = %q, want %q", bill.Description, desc)
	}
}

func TestCreate_Bill_WithSupplier(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Use fixture supplier id=5101
	id, err := repo.Create(ctx, testutil.OwnerID, domain.BillInput{
		SupplierID: 5101,
		Amount:     money.Amount(35000),
		Currency:   "USD",
		DueDate:    "2026-10-31",
		Status:     "pending",
	})
	if err != nil {
		t.Fatalf("Create with supplier: %v", err)
	}

	bill, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if bill.SupplierID != 5101 {
		t.Errorf("SupplierID = %d, want 5101", bill.SupplierID)
	}
}

func TestUpdate_Bill_NoopUpdateSucceeds(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BillInput{
		Amount:      money.Amount(17500),
		Currency:    "USD",
		DueDate:     "2026-12-15",
		Status:      "pending",
		Description: "unchanged retry",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	// A retry that re-sends the stored values changes no column, so MySQL
	// reports zero affected rows even though the row matched.
	if err := repo.Update(ctx, testutil.OwnerID, id, domain.BillInput{
		Amount:      money.Amount(17500),
		Currency:    "USD",
		DueDate:     "2026-12-15",
		Status:      "pending",
		Description: "unchanged retry",
	}); err != nil {
		t.Fatalf("no-op Update must not report an existing bill as not found: %v", err)
	}
}

func TestUpdate_Bill_OmittedSupplierFieldsKeepStoredValues(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	purchaseOrderID := int64(981234)
	id, err := repo.Create(ctx, testutil.OwnerID, domain.BillInput{
		SupplierID:      5101,
		SupplierName:    "Acme Supply",
		PurchaseOrderID: &purchaseOrderID,
		Amount:          money.Amount(24000),
		Currency:        "USD",
		DueDate:         "2026-10-20",
		Status:          "pending",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	// The partial update body omits supplier_id, supplier_name and
	// purchase_order_id; those keep the stored values.
	if err := repo.Update(ctx, testutil.OwnerID, id, domain.BillInput{
		Amount:   money.Amount(24000),
		Currency: "USD",
		DueDate:  "2026-10-20",
		Status:   "pending",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	bill, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get after update: %v", err)
	}
	if bill.SupplierID != 5101 {
		t.Errorf("SupplierID = %d, want 5101 (stored link cleared)", bill.SupplierID)
	}
	if bill.SupplierName != "Acme Supply" {
		t.Errorf("SupplierName = %q, want %q (stored snapshot cleared)", bill.SupplierName, "Acme Supply")
	}
	if bill.PurchaseOrderID == nil || *bill.PurchaseOrderID != purchaseOrderID {
		t.Errorf("PurchaseOrderID = %v, want %d (stored link cleared)", bill.PurchaseOrderID, purchaseOrderID)
	}
}

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
