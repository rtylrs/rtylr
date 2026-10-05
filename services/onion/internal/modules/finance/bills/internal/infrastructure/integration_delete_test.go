package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/finance/bills/internal/domain"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/testutil"
)

func TestDelete_Bill_WrongOwnerReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredBillRepository(db)
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
	repo := configuredBillRepository(db)
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
	repo := configuredBillRepository(db)
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
	repo := configuredBillRepository(db)
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
	repo := configuredBillRepository(db)
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
	repo := configuredBillRepository(db)
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
	repo := configuredBillRepository(db)
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
	repo := configuredBillRepository(db)
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
	repo := configuredBillRepository(db)
	ctx := context.Background()

	purchaseOrderID := insertRound3PurchaseOrder(t, db, testutil.OwnerID, "round3-merge-po")
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
	t.Cleanup(func() {
		_ = repo.Delete(context.Background(), testutil.OwnerID, id)
		_, _ = db.ExecContext(context.Background(), `DELETE FROM erp_purchase_orders WHERE id = ?`, purchaseOrderID)
	})

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
