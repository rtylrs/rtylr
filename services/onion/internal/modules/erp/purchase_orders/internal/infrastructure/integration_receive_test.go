package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/purchase_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/purchase_orders/internal/infrastructure"
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

	var receipts int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM erp_purchase_order_receipts WHERE business_id = ? AND purchase_order_id = ?",
		testutil.OwnerID, poID,
	).Scan(&receipts); err != nil {
		t.Fatalf("count purchase order receipts: %v", err)
	}
	if receipts != 0 {
		t.Errorf("rejected receive wrote %d receipts, want 0", receipts)
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

// TestReceivePurchaseOrder_SecondPartialDeliveryAddsStock pins the fix for a
// second partial delivery being dropped: the ledger allows one movement per
// reference and product, and both deliveries used to reference the purchase
// order, so the second one was skipped while the line was still marked
// received. Each delivery is now its own receipt.
func TestReceivePurchaseOrder_SecondPartialDeliveryAddsStock(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	productID := newReceiveTestProduct(t, db, "partial-deliveries", 1.00)
	poID, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-RECEIVE-TWICE-001",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: productID, QuantityOrdered: 10, UnitCost: 3.0},
		},
	}, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}
	items, err := repo.GetPurchaseOrderItems(ctx, poID)
	if err != nil {
		t.Fatalf("GetPurchaseOrderItems: %v", err)
	}

	receive := func(quantity float64) {
		t.Helper()
		if err := repo.ReceivePurchaseOrder(ctx, testutil.OwnerID, poID, testutil.OwnerID, []domain.PurchaseOrderItem{{
			ID:               items[0].ID,
			PurchaseOrderID:  poID,
			ProductID:        productID,
			QuantityOrdered:  10,
			QuantityReceived: quantity,
		}}); err != nil {
			t.Fatalf("ReceivePurchaseOrder(%v): %v", quantity, err)
		}
	}
	receive(4)
	if got := stockOnHand(t, db, productID); got != 4 {
		t.Fatalf("on_hand after first delivery = %v, want 4", got)
	}
	receive(10)
	if got := stockOnHand(t, db, productID); got != 10 {
		t.Fatalf("on_hand after second delivery = %v, want 10", got)
	}

	status, err := repo.GetPurchaseOrderStatus(ctx, testutil.OwnerID, poID)
	if err != nil {
		t.Fatalf("GetPurchaseOrderStatus: %v", err)
	}
	if status != string(domain.POStatusReceived) {
		t.Errorf("status = %q, want received", status)
	}

	var movements int
	var moved float64
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(quantity_change), 0)
		FROM stock_ledger
		WHERE owner_id = ? AND product_id = ? AND entry_type = 'purchase'`,
		testutil.OwnerID, productID,
	).Scan(&movements, &moved); err != nil {
		t.Fatalf("read purchase movements: %v", err)
	}
	if movements != 2 || moved != 10 {
		t.Errorf("ledger has %d purchase movements totalling %v, want 2 totalling 10", movements, moved)
	}
	if quantity, value := costLayerTotals(t, db, productID); quantity != 10 || value != 30 {
		t.Errorf("cost layers hold %v units worth %v, want 10 worth 30", quantity, value)
	}
}

// TestReceivePurchaseOrder_RepeatedProductLinesAllArrive covers a purchase
// order that lists the same product twice: both lines reach stock, valued at
// their weighted average cost.
func TestReceivePurchaseOrder_RepeatedProductLinesAllArrive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	productID := newReceiveTestProduct(t, db, "repeated-lines", 1.00)
	poID, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-RECEIVE-REPEATED-001",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: productID, QuantityOrdered: 2, UnitCost: 1.0},
			{ProductID: productID, QuantityOrdered: 3, UnitCost: 2.0},
		},
	}, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}
	if err := repo.ReceivePurchaseOrder(ctx, testutil.OwnerID, poID, testutil.OwnerID, nil); err != nil {
		t.Fatalf("ReceivePurchaseOrder: %v", err)
	}

	if got := stockOnHand(t, db, productID); got != 5 {
		t.Errorf("on_hand = %v, want 5", got)
	}
	if quantity, value := costLayerTotals(t, db, productID); quantity != 5 || value != 8 {
		t.Errorf("cost layers hold %v units worth %v, want 5 worth 8", quantity, value)
	}
}

func TestUpdatePurchaseOrderStatus_RefusesStatusesThatMoveStock(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	productID := newReceiveTestProduct(t, db, "status-only-receive", 1.00)
	poID, err := repo.CreatePurchaseOrder(ctx, testutil.OwnerID, domain.PurchaseOrderInput{
		SupplierID: fixtureSupplierID,
		PONumber:   "PO-STATUS-ONLY-001",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: productID, QuantityOrdered: 1, UnitCost: 5.0},
		},
	}, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}
	if err := repo.ApprovePurchaseOrder(ctx, testutil.OwnerID, poID); err != nil {
		t.Fatalf("ApprovePurchaseOrder: %v", err)
	}
	for _, status := range []domain.PurchaseOrderStatus{domain.POStatusPartial, domain.POStatusReceived} {
		err := repo.UpdatePurchaseOrderStatus(ctx, testutil.OwnerID, poID, string(status))
		if !errors.Is(err, domain.ErrPurchaseOrderInvalidTransition) {
			t.Fatalf("UpdatePurchaseOrderStatus(%s) error = %v, want ErrPurchaseOrderInvalidTransition", status, err)
		}
	}
	if got := stockOnHand(t, db, productID); got != 0 {
		t.Errorf("status-only update moved stock to %v", got)
	}
}

// newReceiveTestProduct creates a tracked product of its own so assertions on
// stock and cost layers are not shared with other tests.
func newReceiveTestProduct(t *testing.T, db *sql.DB, name string, costPrice float64) int64 {
	t.Helper()
	ctx := context.Background()
	result, err := db.ExecContext(ctx, `
		INSERT INTO erp_products (owner_id, business_id, supplier_id, name, cost_price, track_inventory)
		VALUES (?, ?, ?, ?, ?, TRUE)`,
		testutil.OwnerID, testutil.OwnerID, fixtureSupplierID, name, costPrice)
	if err != nil {
		t.Fatalf("insert product: %v", err)
	}
	productID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read product id: %v", err)
	}
	t.Cleanup(func() {
		for _, query := range []string{
			"DELETE FROM inventory_cost_layers WHERE owner_id = ? AND product_id = ?",
			"DELETE FROM stock_ledger WHERE owner_id = ? AND product_id = ?",
			"DELETE FROM stock_levels WHERE owner_id = ? AND product_id = ?",
			"DELETE FROM erp_inventory_transactions WHERE owner_id = ? AND product_id = ?",
		} {
			_, _ = db.ExecContext(ctx, query, testutil.OwnerID, productID)
		}
	})
	return productID
}

func costLayerTotals(t *testing.T, db *sql.DB, productID int64) (quantity, value float64) {
	t.Helper()
	if err := db.QueryRowContext(context.Background(), `
		SELECT COALESCE(SUM(quantity_remaining), 0), COALESCE(SUM(quantity_remaining * unit_cost), 0)
		FROM inventory_cost_layers
		WHERE owner_id = ? AND product_id = ?`,
		testutil.OwnerID, productID,
	).Scan(&quantity, &value); err != nil {
		t.Fatalf("read cost layers: %v", err)
	}
	return quantity, value
}
