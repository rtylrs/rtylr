package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/purchase_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/purchase_orders/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestApprovePurchaseOrder_ChangesStatusToApproved(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-APPROVE-001",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: 5201, QuantityOrdered: 2, UnitCost: 8.0},
		},
	}, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}

	if err := repo.ApprovePurchaseOrder(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("ApprovePurchaseOrder: %v", err)
	}

	status, err := repo.GetPurchaseOrderStatus(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetPurchaseOrderStatus: %v", err)
	}
	if status != string(domain.POStatusApproved) {
		t.Errorf("expected status=approved, got %q", status)
	}
}

func TestUpdatePurchaseOrderStatus_SetsReceived(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-RECEIVED-001",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: 5201, QuantityOrdered: 1, UnitCost: 5.0},
		},
	}, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}

	if err := repo.UpdatePurchaseOrderStatus(ctx, testutil.OwnerID, id, string(domain.POStatusReceived)); err != nil {
		t.Fatalf("UpdatePurchaseOrderStatus: %v", err)
	}

	status, err := repo.GetPurchaseOrderStatus(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetPurchaseOrderStatus: %v", err)
	}
	if status != string(domain.POStatusReceived) {
		t.Errorf("expected status=received, got %q", status)
	}
}

func TestUpdatePurchaseOrderStatus_ReceivedIsTerminal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-RECEIVED-TERMINAL-001",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: 5201, QuantityOrdered: 1, UnitCost: 5.0},
		},
	}, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}

	if err := repo.UpdatePurchaseOrderStatus(ctx, testutil.OwnerID, id, string(domain.POStatusReceived)); err != nil {
		t.Fatalf("UpdatePurchaseOrderStatus(received): %v", err)
	}

	err = repo.UpdatePurchaseOrderStatus(ctx, testutil.OwnerID, id, string(domain.POStatusDraft))
	if !errors.Is(err, domain.ErrPurchaseOrderAlreadyReceived) {
		t.Fatalf("reopening received purchase order error = %v, want ErrPurchaseOrderAlreadyReceived", err)
	}

	status, err := repo.GetPurchaseOrderStatus(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetPurchaseOrderStatus: %v", err)
	}
	if status != string(domain.POStatusReceived) {
		t.Fatalf("status after rejected reopen = %q, want received", status)
	}
}

func TestApprovePurchaseOrder_ReceivedIsTerminal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-RECEIVED-TERMINAL-002",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: 5201, QuantityOrdered: 1, UnitCost: 5.0},
		},
	}, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}

	if err := repo.UpdatePurchaseOrderStatus(ctx, testutil.OwnerID, id, string(domain.POStatusReceived)); err != nil {
		t.Fatalf("UpdatePurchaseOrderStatus(received): %v", err)
	}

	err = repo.ApprovePurchaseOrder(ctx, testutil.OwnerID, id)
	if !errors.Is(err, domain.ErrPurchaseOrderAlreadyReceived) {
		t.Fatalf("approving received purchase order error = %v, want ErrPurchaseOrderAlreadyReceived", err)
	}
}

func TestGetPurchaseOrderItems_AfterCreate_ReturnsItems(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-ITEMS-CHECK-001",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: 5201, QuantityOrdered: 3, UnitCost: 7.0},
			{ProductID: 5203, QuantityOrdered: 1, UnitCost: 2.0},
		},
	}, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}

	items, err := repo.GetPurchaseOrderItems(ctx, id)
	if err != nil {
		t.Fatalf("GetPurchaseOrderItems: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestListPurchaseOrders_StatusFilter_Pending(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Create a draft PO (newly created POs are draft)
	_, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-DRAFT-FILTER-001",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: 5201, QuantityOrdered: 1, UnitCost: 1.0},
		},
	}, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}

	orders, err := repo.ListPurchaseOrders(ctx, testutil.OwnerID, string(domain.POStatusDraft), 50, 0)
	if err != nil {
		t.Fatalf("ListPurchaseOrders(status=draft): %v", err)
	}
	for _, o := range orders {
		if o.Status != domain.POStatusDraft {
			t.Errorf("expected status=draft, got %q (id=%d)", o.Status, o.ID)
		}
	}
}

func TestListPurchaseOrders_ScopedToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-INT-SCOPE-001",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: 5201, QuantityOrdered: 1, UnitCost: 1.0},
		},
	}, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}

	other, err := repo.ListPurchaseOrders(ctx, testutil.OwnerID+9999, "", 50, 0)
	if err != nil {
		t.Fatalf("ListPurchaseOrders(other owner): %v", err)
	}
	for _, o := range other {
		if o.OwnerID == testutil.OwnerID {
			t.Errorf("other-owner list contains ownerID=%d record", testutil.OwnerID)
		}
	}
}
