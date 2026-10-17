package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/erp/internal/stock"
	"github.com/voxire/onion/internal/modules/erp/onlineorders/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/onlineorders/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestStoreOrders_Create_CustomerNameStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	const want = "Jane Doe"
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		CustomerName: want, Status: "pending", Total: 5.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	detail, err := repo.GetDetail(context.Background(), testutil.BusinessID, id)
	if err != nil {
		t.Fatalf("GetDetail: %v", err)
	}
	if detail.Order.CustomerName != want {
		t.Errorf("CustomerName = %q, want %q", detail.Order.CustomerName, want)
	}
}

func TestStoreOrders_Create_StatusStartsPending(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	const want = "pending"
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		Status: want, Total: 8.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	detail, err := repo.GetDetail(context.Background(), testutil.BusinessID, id)
	if err != nil {
		t.Fatalf("GetDetail: %v", err)
	}
	if detail.Order.Status != want {
		t.Errorf("Status = %q, want %q", detail.Order.Status, want)
	}
}

func TestStoreOrders_GetDetail_OwnerIDMatches(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		Status: "pending", Total: 10.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	detail, err := repo.GetDetail(context.Background(), testutil.BusinessID, id)
	if err != nil {
		t.Fatalf("GetDetail: %v", err)
	}
	if detail.Order.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", detail.Order.OwnerID, testutil.OwnerID)
	}
}

func TestStoreOrders_Create_MultipleItems(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		Status:   "pending",
		Subtotal: 70.0,
		Total:    70.0,
		Items: []domain.OrderItemInput{
			{Name: "ItemA", Quantity: 2, Price: 20.0, Total: 40.0},
			{Name: "ItemB", Quantity: 1, Price: 30.0, Total: 30.0},
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	detail, err := repo.GetDetail(context.Background(), testutil.BusinessID, id)
	if err != nil {
		t.Fatalf("GetDetail: %v", err)
	}
	if len(detail.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(detail.Items))
	}
}

func TestStoreOrders_Update_ChangesCustomerEmail(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		CustomerEmail: "old@example.com", Status: "pending", Total: 5.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.StoreOrderInput{
		CustomerEmail: "new@example.com", Status: "pending", Total: 5.0,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	detail, err := repo.GetDetail(context.Background(), testutil.BusinessID, id)
	if err != nil {
		t.Fatalf("GetDetail: %v", err)
	}
	if detail.Order.CustomerEmail != "new@example.com" {
		t.Errorf("CustomerEmail = %q, want new@example.com", detail.Order.CustomerEmail)
	}
}

func TestStoreOrders_Update_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		Status: "pending", Total: 3.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = repo.Update(context.Background(), testutil.BusinessID, id, domain.StoreOrderInput{
		Status: "confirmed", Total: 3.0,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestStoreOrders_List_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	before, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	_, err = repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		Status: "pending", Total: 1.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	after, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Errorf("expected count +1: before=%d after=%d", len(before), len(after))
	}
}

func TestStoreOrders_Create_IDsAreUnique(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id1, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{Status: "pending", Total: 1.0})
	if err != nil {
		t.Fatalf("Create A: %v", err)
	}
	id2, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{Status: "pending", Total: 2.0})
	if err != nil {
		t.Fatalf("Create B: %v", err)
	}
	if id1 == id2 {
		t.Errorf("expected unique IDs, both got %d", id1)
	}
}

func TestStoreOrders_Create_ZeroOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for owner_id=0, got %d", len(items))
	}
}

func TestStoreOrders_GetDetail_NonExistent_Fails(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.GetDetail(context.Background(), testutil.BusinessID, 9999999)
	if err == nil {
		t.Error("expected error for non-existent order, got nil")
	}
}

func TestStoreOrders_Create_TotalIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	const want = 123.45
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		Status: "pending", Subtotal: 100.0, Tax: 23.45, Total: want,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	detail, err := repo.GetDetail(context.Background(), testutil.BusinessID, id)
	if err != nil {
		t.Fatalf("GetDetail: %v", err)
	}
	if detail.Order.Total != want {
		t.Errorf("Total = %v, want %v", detail.Order.Total, want)
	}
}

// Regression: confirming an order deducts stock and consumes cost layers, but
// the transition back to pending was accepted without reversing that movement.
// The order then read pending while stock stayed deducted, and the next
// confirmation deducted the same quantity a second time.
func TestStoreOrders_ReopenThenReconfirmDeductsStockOnce(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const onHand = 3.0
	const unitCost = 4.0
	productID, locationID := seedStoreOrderStock(t, db, onHand, unitCost)

	orderID, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		CustomerName: "Store Order Reopen Fixture",
		Status:       "pending",
		Items: []domain.OrderItemInput{
			{ProductID: productID, Quantity: 1},
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM store_order_items WHERE order_id = ?", orderID)
		_, _ = db.ExecContext(ctx, "DELETE FROM store_orders WHERE id = ?", orderID)
	})

	if err := repo.Update(ctx, testutil.BusinessID, orderID, domain.StoreOrderInput{Status: "confirmed"}); err != nil {
		t.Fatalf("confirm order: %v", err)
	}

	// Reopening a confirmed order would leave it pending with the stock still
	// deducted, so the transition must be rejected instead of silently applied.
	if err := repo.Update(ctx, testutil.BusinessID, orderID, domain.StoreOrderInput{Status: "pending"}); !errors.Is(err, domain.ErrInvalidStatus) {
		t.Fatalf("confirmed -> pending error = %v, want ErrInvalidStatus", err)
	}

	// A repeat confirmation is a no-op rather than a second deduction.
	if err := repo.Update(ctx, testutil.BusinessID, orderID, domain.StoreOrderInput{Status: "confirmed"}); err != nil {
		t.Fatalf("re-confirm order: %v", err)
	}

	var status string
	if err := db.QueryRowContext(ctx,
		"SELECT status FROM store_orders WHERE id = ? AND business_id = ?", orderID, testutil.BusinessID,
	).Scan(&status); err != nil {
		t.Fatalf("read order status: %v", err)
	}
	if status != "confirmed" {
		t.Fatalf("order status = %q, want confirmed", status)
	}

	var sales int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM stock_ledger
		WHERE owner_id = ? AND product_id = ? AND reference_type = 'store_order'
		  AND reference_id = ? AND entry_type = 'sale'`,
		testutil.OwnerID, productID, orderID,
	).Scan(&sales); err != nil {
		t.Fatalf("count sale movements: %v", err)
	}
	if sales != 1 {
		t.Fatalf("sale movements for order %d = %d, want exactly 1", orderID, sales)
	}

	got := storeOrderStockOnHand(t, db, productID, locationID)
	if got != onHand-1 {
		t.Errorf("on_hand = %v, want %v (deducted once)", got, onHand-1)
	}

	var layerRemaining float64
	if err := db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(quantity_remaining), 0) FROM inventory_cost_layers
		WHERE owner_id = ? AND product_id = ?`,
		testutil.OwnerID, productID,
	).Scan(&layerRemaining); err != nil {
		t.Fatalf("read cost layers: %v", err)
	}
	if layerRemaining != onHand-1 {
		t.Errorf("cost layer remaining = %v, want %v (consumed once)", layerRemaining, onHand-1)
	}
}

// seedStoreOrderStock creates a dedicated inventory-tracking product with
// on-hand stock and a matching cost layer, so the shared fixture products keep
// their balances. Everything it creates is removed by t.Cleanup.
func seedStoreOrderStock(t *testing.T, db *sql.DB, onHand, unitCost float64) (productID, locationID int64) {
	t.Helper()
	ctx := context.Background()

	result, err := db.ExecContext(ctx, `
		INSERT INTO erp_products (owner_id, business_id, name, sku, cost_price, sale_price, tax_rate, is_active, track_inventory)
		VALUES (?, ?, 'Store Order Reopen Fixture', ?, ?, 10.00, 0, TRUE, TRUE)`,
		testutil.OwnerID, testutil.BusinessID, fmt.Sprintf("store-order-reopen-%d", time.Now().UnixNano()), unitCost)
	if err != nil {
		t.Fatalf("seed product: %v", err)
	}
	productID, err = result.LastInsertId()
	if err != nil {
		t.Fatalf("seed product id: %v", err)
	}

	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM stock_ledger WHERE owner_id = ? AND product_id = ?", testutil.OwnerID, productID)
		_, _ = db.ExecContext(ctx, "DELETE FROM inventory_cost_layers WHERE owner_id = ? AND product_id = ?", testutil.OwnerID, productID)
		_, _ = db.ExecContext(ctx, "DELETE FROM stock_levels WHERE owner_id = ? AND product_id = ?", testutil.OwnerID, productID)
		_, _ = db.ExecContext(ctx, "DELETE FROM erp_products WHERE id = ?", productID)
	})

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin stock seed: %v", err)
	}
	defer tx.Rollback()

	locationID, err = stock.DefaultLocationID(ctx, tx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("default stock location: %v", err)
	}
	if _, _, err := stock.SetTotalTx(ctx, tx, stock.Movement{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		LocationID:    locationID,
		ProductID:     productID,
		EntryType:     stock.EntryAdjustment,
		ReferenceType: "test_seed",
		Note:          "store order reopen test seed",
	}, onHand); err != nil {
		t.Fatalf("seed stock: %v", err)
	}
	// SetTotalTx opens the matching cost layer at the product's cost price.
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit stock seed: %v", err)
	}

	return productID, locationID
}

func storeOrderStockOnHand(t *testing.T, db *sql.DB, productID, locationID int64) float64 {
	t.Helper()
	var onHand float64
	if err := db.QueryRowContext(context.Background(), `
		SELECT COALESCE(SUM(on_hand), 0) FROM stock_levels
		WHERE owner_id = ? AND product_id = ? AND location_id = ?`,
		testutil.OwnerID, productID, locationID,
	).Scan(&onHand); err != nil {
		t.Fatalf("read stock on_hand: %v", err)
	}
	return onHand
}
