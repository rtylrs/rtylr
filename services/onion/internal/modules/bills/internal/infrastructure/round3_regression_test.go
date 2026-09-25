package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/bills/internal/domain"
	"github.com/voxire/onion/internal/modules/bills/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/testutil"
)

func TestRound3BillRepositoryPartialUpdatePreservesAmount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	businessID := int64(987654500)

	id, err := repo.Create(ctx, businessID, domain.BillInput{
		Amount:      money.Amount(12345),
		Currency:    "USD",
		DueDate:     "2030-01-01",
		Status:      "pending",
		Description: "keep me",
	})
	if err != nil {
		t.Fatalf("create bill: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, `DELETE FROM finance_bills WHERE business_id = ?`, businessID) })

	if err := repo.Update(ctx, businessID, id, domain.BillInput{Status: "overdue"}); err != nil {
		t.Fatalf("partial update: %v", err)
	}
	bill, err := repo.Get(ctx, businessID, id)
	if err != nil {
		t.Fatalf("get updated bill: %v", err)
	}
	if bill.Amount != money.Amount(12345) || bill.Currency != "USD" || bill.DueDate.Format("2006-01-02") != "2030-01-01" || bill.Description != "keep me" {
		t.Fatalf("partial update lost fields: %+v", bill)
	}
}

func TestRound3BillRejectsForeignPurchaseOrder(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	businessID := int64(987654501)
	foreignBusinessID := int64(987654502)

	foreignID := insertRound3PurchaseOrder(t, db, foreignBusinessID, "round3-foreign-po")
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM finance_bills WHERE business_id = ?`, businessID)
		_, _ = db.ExecContext(ctx, `DELETE FROM erp_purchase_orders WHERE id = ?`, foreignID)
	})

	foreign := foreignID
	_, err := repo.Create(ctx, businessID, domain.BillInput{
		PurchaseOrderID: &foreign,
		Amount:          money.Amount(1000),
		Currency:        "USD",
		DueDate:         "2030-01-01",
	})
	if !errors.Is(err, domain.ErrBillReferenceNotFound) {
		t.Fatalf("foreign purchase order error = %v, want ErrBillReferenceNotFound", err)
	}
}

func TestRound3BillUpdateRejectsForeignPurchaseOrder(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	businessID := int64(987654504)
	foreignBusinessID := int64(987654505)
	foreignID := insertRound3PurchaseOrder(t, db, foreignBusinessID, "round3-foreign-update-po")

	id, err := repo.Create(ctx, businessID, domain.BillInput{
		Amount:   money.Amount(1000),
		Currency: "USD",
		DueDate:  "2030-01-01",
	})
	if err != nil {
		t.Fatalf("create bill: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM finance_bills WHERE business_id = ?`, businessID)
		_, _ = db.ExecContext(ctx, `DELETE FROM erp_purchase_orders WHERE id = ?`, foreignID)
	})

	foreign := foreignID
	err = repo.Update(ctx, businessID, id, domain.BillInput{
		PurchaseOrderID: &foreign,
		Amount:          money.Amount(1000),
		Currency:        "USD",
		DueDate:         "2030-01-01",
	})
	if !errors.Is(err, domain.ErrBillReferenceNotFound) {
		t.Fatalf("foreign purchase order update error = %v, want ErrBillReferenceNotFound", err)
	}
	bill, err := repo.Get(ctx, businessID, id)
	if err != nil {
		t.Fatalf("get bill after rejected update: %v", err)
	}
	if bill.PurchaseOrderID != nil {
		t.Fatalf("rejected update persisted foreign purchase order %v", bill.PurchaseOrderID)
	}
}

func TestRound3BillAcceptsPurchaseOrderFromSameBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	businessID := int64(987654503)
	purchaseOrderID := insertRound3PurchaseOrder(t, db, businessID, "round3-local-po")
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM finance_bills WHERE business_id = ?`, businessID)
		_, _ = db.ExecContext(ctx, `DELETE FROM erp_purchase_orders WHERE id = ?`, purchaseOrderID)
	})

	id, err := repo.Create(ctx, businessID, domain.BillInput{
		PurchaseOrderID: &purchaseOrderID,
		Amount:          money.Amount(1000),
		Currency:        "USD",
		DueDate:         "2030-01-01",
	})
	if err != nil {
		t.Fatalf("create bill with local purchase order: %v", err)
	}
	if id <= 0 {
		t.Fatalf("created bill id = %d", id)
	}
}

func insertRound3PurchaseOrder(t *testing.T, db interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}, businessID int64, number string) int64 {
	t.Helper()
	result, err := db.ExecContext(context.Background(), `
		INSERT INTO erp_purchase_orders (owner_id, business_id, supplier_id, po_number)
		VALUES (?, ?, 1, ?)`, testutil.OwnerID, businessID, number)
	if err != nil {
		t.Fatalf("insert purchase order: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("purchase order id: %v", err)
	}
	return id
}
