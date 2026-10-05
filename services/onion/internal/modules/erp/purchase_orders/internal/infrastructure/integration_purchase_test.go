package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/purchase_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/purchase_orders/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetPurchaseOrderStatus_NotFound_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetPurchaseOrderStatus(ctx, testutil.OwnerID, 0)
	if err == nil {
		t.Fatal("GetPurchaseOrderStatus(id=0): expected error, got nil")
	}
}

func TestUpdatePurchaseOrderStatus_UpdatesToCancelled(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-INT-005",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: 5201, QuantityOrdered: 1, UnitCost: 1.0},
		},
	}, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}

	err = repo.UpdatePurchaseOrderStatus(ctx, testutil.OwnerID, id, string(domain.POStatusCancelled))
	if err != nil {
		t.Fatalf("UpdatePurchaseOrderStatus: %v", err)
	}

	status, err := repo.GetPurchaseOrderStatus(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetPurchaseOrderStatus: %v", err)
	}
	if status != string(domain.POStatusCancelled) {
		t.Errorf("expected cancelled, got %q", status)
	}
}

func TestUpdatePurchaseOrderStatus_UnknownID_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	err := repo.UpdatePurchaseOrderStatus(context.Background(), testutil.BusinessID, 999999999, string(domain.POStatusCancelled))
	if !errors.Is(err, domain.ErrPurchaseOrderNotFound) {
		t.Fatalf("UpdatePurchaseOrderStatus unknown ID error = %v, want ErrPurchaseOrderNotFound", err)
	}
}

func TestListPurchaseOrders_FilterByStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Create a PO to set to approved
	id, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-INT-FILT-001",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: 5201, QuantityOrdered: 1, UnitCost: 1.0},
		},
	}, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}

	err = repo.ApprovePurchaseOrder(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("ApprovePurchaseOrder: %v", err)
	}

	orders, err := repo.ListPurchaseOrders(ctx, testutil.OwnerID, string(domain.POStatusApproved), 50, 0)
	if err != nil {
		t.Fatalf("ListPurchaseOrders(status=approved): %v", err)
	}
	for _, o := range orders {
		if o.Status != domain.POStatusApproved {
			t.Errorf("expected status=approved, got %q (id=%d)", o.Status, o.ID)
		}
	}
}

func TestGetPurchaseOrderItems_ReturnedWithOrder(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-INT-ITEMS-001",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: 5201, QuantityOrdered: 2, UnitCost: 10.0},
			{ProductID: 5202, QuantityOrdered: 4, UnitCost: 3.5},
		},
	}

	id, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, input, testutil.OwnerID)
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

func TestListPurchaseOrders_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	orders, err := repo.ListPurchaseOrders(ctx, 0, "", 20, 0)
	if err != nil {
		t.Fatalf("ListPurchaseOrders(ownerID=0): %v", err)
	}
	if len(orders) != 0 {
		t.Fatalf("expected 0 orders for ownerID=0, got %d", len(orders))
	}
}

func TestListPurchaseOrders_AllStatuses(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Empty status string means no filter -- should return all
	orders, err := repo.ListPurchaseOrders(ctx, testutil.OwnerID, "", 100, 0)
	if err != nil {
		t.Fatalf("ListPurchaseOrders(all statuses): %v", err)
	}
	if orders == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
}

func TestCreatePurchaseOrder_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-PERSIST-001",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: 5201, QuantityOrdered: 1, UnitCost: 5.0},
		},
	}, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreatePurchaseOrder_SupplierIDIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-SUPPLIER-001",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: 5201, QuantityOrdered: 1, UnitCost: 5.0},
		},
	}, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}

	po, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if po.SupplierID != fixtureSupplierID {
		t.Errorf("expected SupplierID=%d, got %d", fixtureSupplierID, po.SupplierID)
	}
}

func TestGetByID_NonExistent_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, testutil.OwnerID, 9999999)
	if err == nil {
		t.Fatal("expected error for non-existent PO id=9999999, got nil")
	}
}

func TestGetPurchaseOrderStatus_AfterCreate_IsPending(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-STATUS-DRAFT-001",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: 5201, QuantityOrdered: 1, UnitCost: 10.0},
		},
	}, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}

	status, err := repo.GetPurchaseOrderStatus(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetPurchaseOrderStatus: %v", err)
	}
	// Newly created POs start as "draft"
	if status != string(domain.POStatusDraft) {
		t.Errorf("expected status=draft, got %q", status)
	}
}
