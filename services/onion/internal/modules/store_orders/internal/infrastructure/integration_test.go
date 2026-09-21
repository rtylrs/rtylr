package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/store_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/store_orders/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/stock_core"
	"github.com/voxire/onion/internal/testutil"
)

func TestStoreOrders_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestStoreOrders_List_OwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, o := range items {
		if o.BusinessID != testutil.BusinessID {
			t.Errorf("expected business_id %d, got %d", testutil.BusinessID, o.BusinessID)
		}
	}
}

func TestStoreOrders_Create_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		CustomerEmail: "customer@example.com",
		CustomerName:  "Test Customer",
		Status:        "pending",
		Subtotal:      99.0,
		Tax:           9.9,
		ShippingCost:  5.0,
		Total:         113.9,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestStoreOrders_Create_UsesCustomerSync(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	email := fmt.Sprintf("store-orders-sync-%d@example.com", time.Now().UnixNano())
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), "DELETE FROM crm_customers WHERE business_id = ? AND email = ?", testutil.BusinessID, email); err != nil {
			t.Errorf("cleanup synced customer: %v", err)
		}
	})

	firstID, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		CustomerName:  "  Checkout Customer  ",
		CustomerEmail: "  " + email + "  ",
		CustomerPhone: " +1 555 0100 ",
		Status:        "pending",
		Total:         10.50,
	})
	if err != nil {
		t.Fatalf("Create first order: %v", err)
	}
	if firstID <= 0 {
		t.Fatalf("expected positive first order ID, got %d", firstID)
	}

	secondID, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		CustomerName:  "Updated Customer",
		CustomerEmail: email,
		CustomerPhone: "+1 555 0100",
		Status:        "pending",
		Total:         4.25,
	})
	if err != nil {
		t.Fatalf("Create second order: %v", err)
	}
	if secondID <= 0 {
		t.Fatalf("expected positive second order ID, got %d", secondID)
	}

	var firstName, lastName, normalizedEmail, normalizedPhone, totalSpent string
	var visitCount int32
	var businessID int64
	err = db.QueryRowContext(context.Background(), `
		SELECT first_name, COALESCE(last_name, ''), email, COALESCE(phone, ''), total_spent, visit_count, business_id
		FROM crm_customers
		WHERE business_id = ? AND email = ?
		ORDER BY id DESC
		LIMIT 1`, testutil.BusinessID, email).Scan(
		&firstName, &lastName, &normalizedEmail, &normalizedPhone, &totalSpent, &visitCount, &businessID,
	)
	if err != nil {
		t.Fatalf("read synced customer: %v", err)
	}
	if firstName != "Updated" || lastName != "Customer" {
		t.Fatalf("expected canonical updated name, got %q %q", firstName, lastName)
	}
	if normalizedEmail != email || normalizedPhone != "+1 555 0100" {
		t.Fatalf("expected normalized contact, got email=%q phone=%q", normalizedEmail, normalizedPhone)
	}
	if totalSpent != "14.75" || visitCount != 2 {
		t.Fatalf("expected accumulated spend and visits, got total=%q visits=%d", totalSpent, visitCount)
	}
	if businessID != testutil.BusinessID {
		t.Fatalf("expected customer business_id %d, got %d", testutil.BusinessID, businessID)
	}
}

func TestStoreOrders_Create_WithItems(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	input := domain.StoreOrderInput{
		CustomerEmail: "items@example.com",
		Status:        "pending",
		Subtotal:      50.0,
		Tax:           5.0,
		Total:         55.0,
		Items: []domain.OrderItemInput{
			{Name: "Widget", Quantity: 2, Price: 25.0, Total: 50.0},
		},
	}
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, input)
	if err != nil {
		t.Fatalf("Create with items: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestStoreOrders_GetDetail_ReturnsOrder(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		CustomerName: "Detail Test",
		Status:       "pending",
		Subtotal:     10.0,
		Tax:          1.0,
		Total:        11.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	detail, err := repo.GetDetail(context.Background(), testutil.BusinessID, id)
	if err != nil {
		t.Fatalf("GetDetail: %v", err)
	}
	if detail.Order.ID != id {
		t.Errorf("expected order ID %d, got %d", id, detail.Order.ID)
	}
}

func TestStoreOrders_GetDetail_IncludesItems(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		Status:   "pending",
		Subtotal: 30.0,
		Total:    30.0,
		Items: []domain.OrderItemInput{
			{Name: "Gadget", Quantity: 1, Price: 30.0, Total: 30.0},
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	detail, err := repo.GetDetail(context.Background(), testutil.BusinessID, id)
	if err != nil {
		t.Fatalf("GetDetail: %v", err)
	}
	if len(detail.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(detail.Items))
	}
	if detail.Items[0].Name != "Gadget" {
		t.Errorf("expected item name Gadget, got %s", detail.Items[0].Name)
	}
}

func TestStoreOrders_Update_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		Status: "pending", Subtotal: 5.0, Total: 5.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.StoreOrderInput{
		Status: "confirmed", Subtotal: 5.0, Total: 5.0,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	detail, err := repo.GetDetail(context.Background(), testutil.BusinessID, id)
	if err != nil {
		t.Fatalf("GetDetail after Update: %v", err)
	}
	if detail.Order.Status != "confirmed" {
		t.Errorf("expected status confirmed, got %s", detail.Order.Status)
	}
}

func TestStoreOrders_List_UnknownOwner_Empty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), 9999999)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 for unknown owner, got %d", len(items))
	}
}

func TestStoreOrders_Create_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		CustomerEmail: "appears@example.com",
		Status:        "pending",
		Total:         15.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, o := range items {
		if o.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("created order id=%d not found in list", id)
	}
}

func TestStoreOrders_GetDetail_CrossOwner_Fails(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		Status: "pending", Total: 5.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = repo.GetDetail(context.Background(), 9999999, id)
	if err == nil {
		t.Error("expected error when fetching order with wrong owner")
	}
}

func TestStoreOrders_Create_WithShippingAddress(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		Status:          "pending",
		Subtotal:        25.0,
		Total:           25.0,
		ShippingAddress: `{"street":"123 Main St","city":"Testville"}`,
	})
	if err != nil {
		t.Fatalf("Create with shipping address: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestStoreOrders_List_IsNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestStoreOrders_Create_CustomerEmailStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	const want = "email-stored@example.com"
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		CustomerEmail: want, Status: "pending", Total: 10.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	detail, err := repo.GetDetail(context.Background(), testutil.BusinessID, id)
	if err != nil {
		t.Fatalf("GetDetail: %v", err)
	}
	if detail.Order.CustomerEmail != want {
		t.Errorf("CustomerEmail = %q, want %q", detail.Order.CustomerEmail, want)
	}
}

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

	locationID, err = stock_core.DefaultLocationID(ctx, tx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("default stock location: %v", err)
	}
	if _, _, err := stock_core.SetTotalTx(ctx, tx, stock_core.Movement{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		LocationID:    locationID,
		ProductID:     productID,
		EntryType:     stock_core.EntryAdjustment,
		ReferenceType: "test_seed",
		Note:          "store order reopen test seed",
	}, onHand); err != nil {
		t.Fatalf("seed stock: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit stock seed: %v", err)
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO inventory_cost_layers (owner_id, business_id, product_id, location_id, quantity_remaining, unit_cost, reference_type)
		VALUES (?, ?, ?, ?, ?, ?, 'test_seed')`,
		testutil.OwnerID, testutil.BusinessID, productID, locationID, onHand, unitCost); err != nil {
		t.Fatalf("seed cost layer: %v", err)
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
