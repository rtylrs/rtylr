package infrastructure_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/purchase_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/purchase_orders/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const fixtureSupplierID int64 = 5101

func TestListPurchaseOrders_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	orders, err := repo.ListPurchaseOrders(ctx, testutil.OwnerID, "", 20, 0)
	if err != nil {
		t.Fatalf("ListPurchaseOrders: unexpected error: %v", err)
	}
	if orders == nil {
		t.Fatal("ListPurchaseOrders: expected non-nil slice, got nil")
	}
}

func TestListPurchaseOrders_WrongOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	orders, err := repo.ListPurchaseOrders(ctx, 0, "", 20, 0)
	if err != nil {
		t.Fatalf("ListPurchaseOrders(ownerID=0): unexpected error: %v", err)
	}
	if len(orders) != 0 {
		t.Fatalf("ListPurchaseOrders(ownerID=0): expected 0 rows, got %d", len(orders))
	}
}

func TestCreatePurchaseOrder_ReturnsInsertedID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-INT-001",
		TaxAmount:  5.0,
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: 5201, QuantityOrdered: 2, UnitCost: 10.0},
		},
	}

	id, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, input, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: unexpected error: %v", err)
	}
	if id <= 0 {
		t.Fatalf("CreatePurchaseOrder: expected positive ID, got %d", id)
	}
}

func TestCreatePurchaseOrder_RejectsUnscopedReferences(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	_, err := repo.CreatePurchaseOrder(context.Background(), testutil.OwnerID, domain.PurchaseOrderInput{
		SupplierID: 999999999,
		PONumber:   "PO-CROSS-BUSINESS-001",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: 5201, QuantityOrdered: 1, UnitCost: 5},
		},
	}, testutil.OwnerID)
	if !errors.Is(err, domain.ErrPurchaseOrderReferenceNotFound) {
		t.Fatalf("CreatePurchaseOrder cross-business supplier error = %v, want ErrPurchaseOrderReferenceNotFound", err)
	}
}

// Regression: the reference-validation failure returned from an err shadowed
// inside the if-statement, so the deferred rollback never ran and the
// transaction's pooled connection was never released. Pinning the pool to a
// single connection turns a leak into a blocked (and then context-expired)
// second call.
func TestCreatePurchaseOrder_ValidationFailureReleasesConnection(t *testing.T) {
	db := testutil.OpenDB(t)
	db.SetMaxOpenConns(1)
	repo := infrastructure.NewRepository(db)

	input := domain.PurchaseOrderInput{
		SupplierID: 999999999,
		PONumber:   "PO-LEAK-001",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: 5201, QuantityOrdered: 1, UnitCost: 5},
		},
	}

	for attempt := 1; attempt <= 2; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, input, testutil.OwnerID)
		cancel()

		if !errors.Is(err, domain.ErrPurchaseOrderReferenceNotFound) {
			t.Fatalf("attempt %d: error = %v, want ErrPurchaseOrderReferenceNotFound (a leaked transaction shows up as a context deadline here)", attempt, err)
		}
	}
}

func TestCreatePurchaseOrder_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-INT-002",
		TaxAmount:  0,
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: 5202, QuantityOrdered: 1, UnitCost: 5.0},
		},
	}

	id, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, input, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}

	orders, err := repo.ListPurchaseOrders(ctx, testutil.OwnerID, "", 100, 0)
	if err != nil {
		t.Fatalf("ListPurchaseOrders: %v", err)
	}

	var found bool
	for _, o := range orders {
		if o.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created purchase order id=%d not found in list", id)
	}
}

func TestGetByID_ExistingPurchaseOrder(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-INT-003",
		TaxAmount:  2.5,
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: 5201, QuantityOrdered: 3, UnitCost: 8.0},
		},
	}

	id, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, input, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}

	po, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID(id=%d): unexpected error: %v", id, err)
	}
	if po.ID != id {
		t.Errorf("GetByID: ID mismatch: got %d, want %d", po.ID, id)
	}
}

func TestGetByID_NotFound_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, testutil.OwnerID, 0)
	if err == nil {
		t.Fatal("GetByID(id=0): expected error for non-existent PO, got nil")
	}
}

func TestGetPurchaseOrderStatus_DefaultIsDraft(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-INT-004",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: 5201, QuantityOrdered: 1, UnitCost: 10.0},
		},
	}

	id, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, input, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}

	status, err := repo.GetPurchaseOrderStatus(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetPurchaseOrderStatus: %v", err)
	}
	if status != "draft" {
		t.Errorf("expected status=draft, got %q", status)
	}
}
