package infrastructure_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/orders/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

func TestListOrders_FilterByStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()

	orders, _, _, err := repo.ListOrders(ctx, testutil.OwnerID, 0, "paid", "", "", "", 500, 0)
	if err != nil {
		t.Fatalf("ListOrders: %v", err)
	}
	for _, o := range orders {
		if string(o.Status) != "paid" {
			t.Errorf("expected status='paid', got %q for order id=%d", o.Status, o.ID)
		}
	}
}

func TestGetOrderItems_FixtureOrder(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()

	items, err := repo.GetOrderItems(ctx, 3301)
	if err != nil {
		t.Fatalf("GetOrderItems: %v", err)
	}
	_ = items
}

func TestUpdateOrderItemStatus_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()
	productID := insertTestProduct(ctx, t, db, "OrderItemStatus", 5.00)

	orderNum := fmt.Sprintf("TEST-ORD-%d", 9004)
	id, err := repo.CreateOrder(ctx, testutil.OwnerID, 4101, orderNum, domain.OrderInput{
		SessionID: 3201,
		Type:      domain.OrderTypeTakeaway,
		Items: []domain.OrderItemInput{
			{MenuItemID: productID, Quantity: 1, UnitPrice: 0},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_order_items WHERE order_id = ?`, id)
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_orders WHERE id = ?`, id)
	})

	items, err := repo.GetOrderItems(ctx, id)
	if err != nil {
		t.Fatalf("GetOrderItems: %v", err)
	}
	if len(items) == 0 {
		t.Skip("no items to update status for")
	}

	if err := repo.UpdateOrderItemStatus(ctx, id, items[0].ID, domain.ItemStatusPreparing); err != nil {
		t.Fatalf("UpdateOrderItemStatus: %v", err)
	}
}

func TestCreateOrder_UsesServerPriceInsteadOfClientPrice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()

	productID := insertTestProduct(ctx, t, db, "ServerPriced", 15.00)
	_, err := db.ExecContext(ctx, `UPDATE erp_products SET tax_rate = 10.00 WHERE id = ?`, productID)
	if err != nil {
		t.Fatalf("set test product tax rate: %v", err)
	}

	orderNum := fmt.Sprintf("TEST-ORD-PRICE-%d", 9005)
	orderID, err := repo.CreateOrder(ctx, testutil.OwnerID, 4101, orderNum, domain.OrderInput{
		Type: domain.OrderTypeTakeaway,
		Items: []domain.OrderItemInput{{
			MenuItemID: productID,
			Quantity:   1,
			UnitPrice:  0,
			TaxRate:    0,
		}},
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_order_items WHERE order_id = ?`, orderID)
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_orders WHERE id = ?`, orderID)
	})

	order, err := repo.GetOrder(ctx, testutil.OwnerID, orderID)
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if order.Total != 16.50 {
		t.Fatalf("order total = %.2f, want 16.50 from server price and tax", order.Total)
	}
	if len(order.Items) != 1 || order.Items[0].UnitPrice != 15.00 || order.Items[0].TaxRate != 10.00 {
		t.Fatalf("stored item pricing = %+v, want unit price 15 and tax 10", order.Items)
	}
}

func TestSplitOrder_SplitsSubtotalAndTax(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()
	productID := insertTestProduct(ctx, t, db, "OrderSplit", 10.00)
	if _, err := db.ExecContext(ctx, `UPDATE erp_products SET tax_rate = 10.00 WHERE id = ?`, productID); err != nil {
		t.Fatalf("set test product tax rate: %v", err)
	}

	orderNum := fmt.Sprintf("TEST-ORD-%d", 9101)
	// Unit price and tax rate are resolved from the catalog, not taken from the
	// caller, so the item inputs carry only the product and the quantity. Both
	// lines resolve to the same product at 10.00 and a 10 percent tax rate.
	id, err := repo.CreateOrder(ctx, testutil.OwnerID, 4101, orderNum, domain.OrderInput{
		SessionID: 3201,
		Type:      domain.OrderTypeTakeaway,
		Items: []domain.OrderItemInput{
			{MenuItemID: productID, Quantity: 1},
			{MenuItemID: productID, Quantity: 1},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	// Registered as soon as the row exists, not after the split. The previous
	// cleanup sat below SplitOrder, so a failure there skipped it and left
	// TEST-ORD-9101 behind; the next run then collided on the order number and
	// reported a duplicate key from CreateOrder, hiding the real error (#1185).
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_order_items WHERE order_id = ?`, id)
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_orders WHERE id = ?`, id)
	})

	order, err := repo.GetOrder(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if len(order.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(order.Items))
	}

	newIDs, err := repo.SplitOrder(ctx, testutil.OwnerID, id, [][]int64{{order.Items[0].ID}})
	if err != nil {
		t.Fatalf("SplitOrder: %v", err)
	}
	if len(newIDs) != 1 {
		t.Fatalf("expected 1 split order, got %d", len(newIDs))
	}
	t.Cleanup(func() {
		for _, oid := range newIDs {
			_, _ = db.ExecContext(ctx, `DELETE FROM pos_order_items WHERE order_id = ?`, oid)
			_, _ = db.ExecContext(ctx, `DELETE FROM pos_orders WHERE id = ?`, oid)
		}
	})

	child, err := repo.GetOrder(ctx, testutil.OwnerID, newIDs[0])
	if err != nil {
		t.Fatalf("GetOrder split child: %v", err)
	}
	if child.Subtotal != 10 {
		t.Errorf("expected child subtotal=10, got %.2f", child.Subtotal)
	}
	if child.TaxAmount != 1 {
		t.Errorf("expected child tax=1, got %.2f", child.TaxAmount)
	}
	if child.Total != 11 {
		t.Errorf("expected child total=11, got %.2f", child.Total)
	}
}

// A split into two groups used to ask for an order number once per group.
// GenerateOrderNumber counts committed rows through the pool, so while the
// split transaction was still open every group was offered the same
// ORD-<date>-NNNN; the second child insert then violated
// uq_pos_orders_number (owner_id, order_number) and the whole split failed
// with a duplicate-key error.
func TestSplitOrder_MultipleGroupsGetDistinctOrderNumbers(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()
	productID := insertTestProduct(ctx, t, db, "OrderSplitTwoGroups", 10.00)

	orderNum := fmt.Sprintf("TEST-ORD-%d", 9201)
	id, err := repo.CreateOrder(ctx, testutil.OwnerID, 4101, orderNum, domain.OrderInput{
		SessionID: 3201,
		Type:      domain.OrderTypeTakeaway,
		Items: []domain.OrderItemInput{
			{MenuItemID: productID, Quantity: 1, UnitPrice: 10, TaxRate: 10},
			{MenuItemID: productID, Quantity: 1, UnitPrice: 20},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_order_items WHERE order_id = ?`, id)
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_orders WHERE id = ?`, id)
	})

	order, err := repo.GetOrder(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if len(order.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(order.Items))
	}

	newIDs, err := repo.SplitOrder(ctx, testutil.OwnerID, id, [][]int64{
		{order.Items[0].ID},
		{order.Items[1].ID},
	})
	if err != nil {
		t.Fatalf("SplitOrder with two groups: %v", err)
	}
	if len(newIDs) != 2 {
		t.Fatalf("expected 2 split orders, got %d", len(newIDs))
	}
	t.Cleanup(func() {
		for _, oid := range newIDs {
			_, _ = db.ExecContext(ctx, `DELETE FROM pos_order_items WHERE order_id = ?`, oid)
			_, _ = db.ExecContext(ctx, `DELETE FROM pos_orders WHERE id = ?`, oid)
		}
	})

	first, err := repo.GetOrder(ctx, testutil.OwnerID, newIDs[0])
	if err != nil {
		t.Fatalf("GetOrder first split child: %v", err)
	}
	second, err := repo.GetOrder(ctx, testutil.OwnerID, newIDs[1])
	if err != nil {
		t.Fatalf("GetOrder second split child: %v", err)
	}
	if first.OrderNumber == "" || second.OrderNumber == "" {
		t.Fatalf("split children have empty order numbers: %q and %q", first.OrderNumber, second.OrderNumber)
	}
	if first.OrderNumber == second.OrderNumber {
		t.Fatalf("split children share order number %q; uq_pos_orders_number rejects the second insert", first.OrderNumber)
	}
}

// ─── Regression: order status writes must reject unknown targets ─────────────

func TestPublicMenuReadsERPCatalogAndEnabledExtension(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()
	productID := insertTestProduct(ctx, t, db, "PublicCatalog", 12.50)
	if _, err := db.Exec("UPDATE erp_products SET tax_rate = 8.00 WHERE id = ?", productID); err != nil {
		t.Fatal(err)
	}
	item, err := repo.GetPublicMenuItem(ctx, testutil.OwnerID, productID)
	if err != nil || item.Name != "PublicCatalog" || item.UnitPrice != 12.50 || item.TaxRate != 8 {
		t.Fatalf("ERP sales product: %+v, %v", item, err)
	}
	if _, err := repo.GetPublicMenuItem(ctx, foreignBusinessID, productID); !errors.Is(err, domain.ErrPublicMenuItemNotFound) {
		t.Fatalf("foreign product access: %v", err)
	}
	if _, err := db.Exec("UPDATE erp_products SET is_active = FALSE WHERE id = ?", productID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetPublicMenuItem(ctx, testutil.OwnerID, productID); !errors.Is(err, domain.ErrPublicMenuItemNotFound) {
		t.Fatalf("inactive product access: %v", err)
	}
	// Use an isolated business ID so no shared fixture grant is changed.
	const grantBusinessID = 997123
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM business_addons WHERE business_id = ?", grantBusinessID) })
	if _, err := db.Exec("INSERT INTO business_addons (business_id, addon_id, is_enabled) VALUES (?, 'online_ordering', TRUE)", grantBusinessID); err != nil {
		t.Fatal(err)
	}
	enabled, err := repo.IsOnlineOrderingEnabled(ctx, grantBusinessID)
	if err != nil || !enabled {
		t.Fatalf("enabled extension: %v, %v", enabled, err)
	}
	if _, err := db.Exec("UPDATE business_addons SET is_enabled = FALSE WHERE business_id = ?", grantBusinessID); err != nil {
		t.Fatal(err)
	}
	enabled, err = repo.IsOnlineOrderingEnabled(ctx, grantBusinessID)
	if err != nil || enabled {
		t.Fatalf("disabled extension: %v, %v", enabled, err)
	}
}
