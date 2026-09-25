package infrastructure_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/voxire/onion/internal/modules/purchase_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/purchase_orders/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// TestReceivePurchaseOrder_RejectsQuantityAboveStoredOrdered pins the fix for a
// receive request that supplies a quantity_received larger than the line's
// stored quantity_ordered (while echoing the same inflated quantity_ordered):
// the stored ordered quantity is the bound, so the request must be rejected and
// leave stock, cost layers, line state, and PO status untouched.
func TestReceivePurchaseOrder_RejectsQuantityAboveStoredOrdered(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const productID int64 = 5201
	poID, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-RECEIVE-BOUND-001",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: productID, QuantityOrdered: 5, UnitCost: 10.0},
		},
	}, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}

	items, err := repo.GetPurchaseOrderItems(ctx, poID)
	if err != nil {
		t.Fatalf("GetPurchaseOrderItems: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 purchase order item, got %d", len(items))
	}

	onHandBefore := stockOnHand(t, db, productID)

	// The submitted quantity_ordered matches quantity_received so the old
	// request-level bound accepted it; the stored line is only 5.
	err = repo.ReceivePurchaseOrder(ctx, testutil.OwnerID, poID, testutil.OwnerID, []domain.PurchaseOrderItem{{
		ID:               items[0].ID,
		PurchaseOrderID:  poID,
		ProductID:        productID,
		QuantityOrdered:  1_000_000,
		QuantityReceived: 1_000_000,
		UnitCost:         10.0,
	}})
	if err == nil {
		t.Fatal("ReceivePurchaseOrder accepted a quantity_received above the stored quantity_ordered")
	}

	if onHandAfter := stockOnHand(t, db, productID); onHandAfter != onHandBefore {
		t.Errorf("rejected receive changed on_hand: got %v, want %v", onHandAfter, onHandBefore)
	}

	var costLayers int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM inventory_cost_layers WHERE business_id = ? AND product_id = ? AND reference_type = 'purchase_order' AND reference_id = ?",
		testutil.OwnerID, productID, poID,
	).Scan(&costLayers); err != nil {
		t.Fatalf("count inventory_cost_layers: %v", err)
	}
	if costLayers != 0 {
		t.Errorf("rejected receive wrote %d inventory_cost_layers rows, want 0", costLayers)
	}

	status, err := repo.GetPurchaseOrderStatus(ctx, testutil.OwnerID, poID)
	if err != nil {
		t.Fatalf("GetPurchaseOrderStatus: %v", err)
	}
	if status != string(domain.POStatusDraft) {
		t.Errorf("rejected receive changed status to %q, want %q", status, domain.POStatusDraft)
	}

	storedItems, err := repo.GetPurchaseOrderItems(ctx, poID)
	if err != nil {
		t.Fatalf("GetPurchaseOrderItems after rejected receive: %v", err)
	}
	if storedItems[0].QuantityReceived != 0 {
		t.Errorf("rejected receive changed quantity_received to %v, want 0", storedItems[0].QuantityReceived)
	}
}

// TestReceivePurchaseOrder_AcceptsPartialQuantityWithinStoredBound keeps the
// legitimate partial receive working: a quantity_received below the stored
// quantity_ordered is accepted, adds that amount to stock, and marks the PO
// partial.
func TestReceivePurchaseOrder_AcceptsPartialQuantityWithinStoredBound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const productID int64 = 5202
	poID, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-RECEIVE-PARTIAL-001",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: productID, QuantityOrdered: 5, UnitCost: 2.0},
		},
	}, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}

	items, err := repo.GetPurchaseOrderItems(ctx, poID)
	if err != nil {
		t.Fatalf("GetPurchaseOrderItems: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 purchase order item, got %d", len(items))
	}

	onHandBefore := stockOnHand(t, db, productID)

	if err := repo.ReceivePurchaseOrder(ctx, testutil.OwnerID, poID, testutil.OwnerID, []domain.PurchaseOrderItem{{
		ID:               items[0].ID,
		PurchaseOrderID:  poID,
		ProductID:        productID,
		QuantityOrdered:  5,
		QuantityReceived: 2,
		UnitCost:         2.0,
	}}); err != nil {
		t.Fatalf("ReceivePurchaseOrder partial: %v", err)
	}

	if onHandAfter := stockOnHand(t, db, productID); onHandAfter != onHandBefore+2 {
		t.Errorf("partial receive on_hand = %v, want %v", onHandAfter, onHandBefore+2)
	}

	status, err := repo.GetPurchaseOrderStatus(ctx, testutil.OwnerID, poID)
	if err != nil {
		t.Fatalf("GetPurchaseOrderStatus: %v", err)
	}
	if status != string(domain.POStatusPartial) {
		t.Errorf("partial receive status = %q, want %q", status, domain.POStatusPartial)
	}

	storedItems, err := repo.GetPurchaseOrderItems(ctx, poID)
	if err != nil {
		t.Fatalf("GetPurchaseOrderItems after partial receive: %v", err)
	}
	if storedItems[0].QuantityReceived != 2 {
		t.Errorf("partial receive quantity_received = %v, want 2", storedItems[0].QuantityReceived)
	}
}

// stockOnHand sums the tenant's on-hand stock for a product across locations.
func stockOnHand(t *testing.T, db *sql.DB, productID int64) float64 {
	t.Helper()
	var onHand float64
	if err := db.QueryRowContext(context.Background(),
		"SELECT COALESCE(SUM(on_hand), 0) FROM stock_levels WHERE owner_id = ? AND product_id = ?",
		testutil.OwnerID, productID,
	).Scan(&onHand); err != nil {
		t.Fatalf("read stock on_hand: %v", err)
	}
	return onHand
}
