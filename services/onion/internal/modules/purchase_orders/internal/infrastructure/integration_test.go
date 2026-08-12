package infrastructure_test

import (
	"context"
	"errors"
	"testing"

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
