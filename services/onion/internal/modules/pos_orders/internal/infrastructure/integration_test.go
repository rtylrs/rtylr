package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const foreignBusinessID int64 = 9002

func insertTestProduct(ctx context.Context, t *testing.T, db *sql.DB, name string, salePrice float64) int64 {
	t.Helper()
	result, err := db.ExecContext(ctx, `
		INSERT INTO erp_products (owner_id, business_id, name, sku, sale_price, is_active)
		VALUES (?, ?, ?, ?, ?, TRUE)
	`, testutil.OwnerID, testutil.OwnerID, name, "TEST-"+name, salePrice)
	if err != nil {
		t.Fatalf("insert test product: %v", err)
	}
	productID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("test product id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM erp_products WHERE id = ?`, productID)
	})
	return productID
}

func TestListLocations_ReturnsFixtureRows(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	locations, err := repo.ListLocations(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locations) == 0 {
		t.Fatal("expected fixture rows, got none")
	}

	var found bool
	for _, loc := range locations {
		if loc.ID == 3001 && loc.Name == "Test Branch" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected location id=3001 named 'Test Branch', got %+v", locations)
	}
}

func TestGetOrder_ReturnsFixturePaidOrder(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	order, err := repo.GetOrder(ctx, testutil.OwnerID, 3301)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order == nil {
		t.Fatal("expected order, got nil")
	}
	if string(order.Status) != "paid" {
		t.Errorf("expected status='paid', got %q", order.Status)
	}
	const wantSubtotal = 150.00
	if order.Subtotal != wantSubtotal {
		t.Errorf("expected subtotal=%.2f, got %.2f", wantSubtotal, order.Subtotal)
	}
}

func TestListOrders_ReturnsBothFixtureOrders(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// sessionID=3201 has 2 orders in fixtures
	orders, _, _, err := repo.ListOrders(ctx, testutil.OwnerID, 3201, "", "", "", "", 500, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orders) < 2 {
		t.Fatalf("expected at least 2 orders, got %d", len(orders))
	}
	foundWithItems := false
	for _, order := range orders {
		if order.ID == 3301 && len(order.Items) > 0 {
			foundWithItems = true
			break
		}
	}
	if !foundWithItems {
		t.Fatalf("expected listed fixture order 3301 to include purchased items, got %+v", orders)
	}
}

func TestCreateLocation_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateLocation(ctx, testutil.OwnerID, domain.LocationInput{
		Name:    "Test Create Location",
		Address: "123 Main St",
		Phone:   "555-0001",
	})
	if err != nil {
		t.Fatalf("CreateLocation: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive id, got %d", id)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_locations WHERE id = ?`, id)
	})
}

func TestCreateLocation_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateLocation(ctx, testutil.OwnerID, domain.LocationInput{
		Name: "Appear In List Location",
	})
	if err != nil {
		t.Fatalf("CreateLocation: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_locations WHERE id = ?`, id)
	})

	locations, err := repo.ListLocations(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListLocations: %v", err)
	}
	found := false
	for _, loc := range locations {
		if loc.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected location id=%d to appear in list", id)
	}
}

func TestListLocations_IsolatedByOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	locations, err := repo.ListLocations(ctx, 9999999)
	if err != nil {
		t.Fatalf("ListLocations: %v", err)
	}
	for _, loc := range locations {
		if loc.OwnerID == testutil.OwnerID {
			t.Errorf("isolation violated: location id=%d belongs to fixture owner", loc.ID)
		}
	}
}

func TestCreateTable_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateTable(ctx, testutil.OwnerID, domain.TableInput{
		LocationID: 3001,
		Name:       "Table T1",
		Capacity:   4,
	})
	if err != nil {
		t.Fatalf("CreateTable: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive id, got %d", id)
	}
	t.Cleanup(func() {
		_ = repo.DeleteTable(ctx, testutil.OwnerID, id)
	})
}

func TestListTables_ForLocation(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateTable(ctx, testutil.OwnerID, domain.TableInput{
		LocationID: 3001,
		Name:       "Filter Location Table",
		Capacity:   2,
	})
	if err != nil {
		t.Fatalf("CreateTable: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.DeleteTable(ctx, testutil.OwnerID, id)
	})

	tables, err := repo.ListTables(ctx, testutil.OwnerID, 3001, "")
	if err != nil {
		t.Fatalf("ListTables: %v", err)
	}
	found := false
	for _, tbl := range tables {
		if tbl.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected table id=%d to appear in list for locationID=3001", id)
	}
}

func TestListTables_FilterByStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateTable(ctx, testutil.OwnerID, domain.TableInput{
		LocationID: 3001,
		Name:       "Status Filter Table",
		Capacity:   2,
	})
	if err != nil {
		t.Fatalf("CreateTable: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.DeleteTable(ctx, testutil.OwnerID, id)
	})

	tables, err := repo.ListTables(ctx, testutil.OwnerID, 0, "available")
	if err != nil {
		t.Fatalf("ListTables: %v", err)
	}
	for _, tbl := range tables {
		if string(tbl.Status) != "available" {
			t.Errorf("expected status='available', got %q for table id=%d", tbl.Status, tbl.ID)
		}
	}
	_ = id
}

func TestUpdateTable_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateTable(ctx, testutil.OwnerID, domain.TableInput{
		LocationID: 3001,
		Name:       "Before Update",
		Capacity:   2,
	})
	if err != nil {
		t.Fatalf("CreateTable: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.DeleteTable(ctx, testutil.OwnerID, id)
	})

	if err := repo.UpdateTable(ctx, testutil.OwnerID, domain.TableInput{
		ID:         id,
		LocationID: 3001,
		Name:       "After Update",
		Capacity:   4,
	}); err != nil {
		t.Fatalf("UpdateTable: %v", err)
	}
}

func TestUpdateTableStatus_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateTable(ctx, testutil.OwnerID, domain.TableInput{
		LocationID: 3001,
		Name:       "Status Change Table",
		Capacity:   2,
	})
	if err != nil {
		t.Fatalf("CreateTable: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.DeleteTable(ctx, testutil.OwnerID, id)
	})

	if err := repo.UpdateTableStatus(ctx, testutil.OwnerID, id, domain.TableStatusOccupied); err != nil {
		t.Fatalf("UpdateTableStatus: %v", err)
	}

	tables, err := repo.ListTables(ctx, testutil.OwnerID, 0, "occupied")
	if err != nil {
		t.Fatalf("ListTables: %v", err)
	}
	found := false
	for _, tbl := range tables {
		if tbl.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected table id=%d to appear with status='occupied'", id)
	}
}

func TestDeleteTable_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateTable(ctx, testutil.OwnerID, domain.TableInput{
		LocationID: 3001,
		Name:       "Delete Me Table",
		Capacity:   2,
	})
	if err != nil {
		t.Fatalf("CreateTable: %v", err)
	}

	if err := repo.DeleteTable(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("DeleteTable: %v", err)
	}

	tables, err := repo.ListTables(ctx, testutil.OwnerID, 0, "")
	if err != nil {
		t.Fatalf("ListTables: %v", err)
	}
	for _, tbl := range tables {
		if tbl.ID == id {
			t.Errorf("expected table id=%d to be deleted", id)
		}
	}
}

func TestOpenSession_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Use a unique staffID to avoid ErrSessionAlreadyOpen from other tests.
	id, err := repo.OpenSession(ctx, testutil.OwnerID, 4199, 3001, 100.00)
	if err != nil {
		t.Fatalf("OpenSession: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive session id, got %d", id)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_sessions WHERE id = ?`, id)
	})
}

func TestGetSessionByID_ReturnsFixtureSession(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	sess, err := repo.GetSessionByID(ctx, testutil.OwnerID, 3201)
	if err != nil {
		t.Fatalf("GetSessionByID: %v", err)
	}
	if sess == nil {
		t.Fatal("expected session, got nil")
	}
	if sess.ID != 3201 {
		t.Errorf("expected session id=3201, got %d", sess.ID)
	}
}

func TestGetSessionByID_UnknownIDReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetSessionByID(ctx, testutil.OwnerID, 9999999)
	if err == nil {
		t.Error("expected error for unknown session id, got nil")
	}
}

func TestListSessions_ReturnsFixtureSession(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	sessions, err := repo.ListSessions(ctx, testutil.OwnerID, 0, "", "")
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	found := false
	for _, s := range sessions {
		if s.ID == 3201 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected fixture session id=3201 in list")
	}
}

func TestListSessions_FilterByLocation(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	sessions, err := repo.ListSessions(ctx, testutil.OwnerID, 3001, "", "")
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	for _, s := range sessions {
		if s.LocationID != 3001 {
			t.Errorf("expected locationID=3001, got %d for session id=%d", s.LocationID, s.ID)
		}
	}
}

func TestGetCurrentOpenSession_ReturnsNilWhenNone(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// staffID=9999999 has no open session in fixtures.
	sess, err := repo.GetCurrentOpenSession(ctx, testutil.OwnerID, 9999999)
	if err != nil {
		t.Fatalf("GetCurrentOpenSession: %v", err)
	}
	if sess != nil {
		t.Errorf("expected nil session for unknown staff, got id=%d", sess.ID)
	}
}

func TestGetSessionTotals_FixtureSession(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	totalSales, cashSales, cardSales, onlineSales, totalOrders, err := repo.GetSessionTotals(ctx, 3201)
	if err != nil {
		t.Fatalf("GetSessionTotals: %v", err)
	}
	if totalSales < 0 || cashSales < 0 || cardSales < 0 || onlineSales < 0 {
		t.Error("expected non-negative totals")
	}
	if totalOrders < 0 {
		t.Errorf("expected non-negative totalOrders, got %d", totalOrders)
	}
}

func TestUpdateSessionTotals_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.OpenSession(ctx, testutil.OwnerID, 4198, 3001, 50.00)
	if err != nil {
		t.Fatalf("OpenSession: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_sessions WHERE id = ?`, id)
	})

	if err := repo.UpdateSessionTotals(ctx, id, 200.00, 100.00, 80.00, 20.00, 5); err != nil {
		t.Fatalf("UpdateSessionTotals: %v", err)
	}
}

func TestCloseSession_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.OpenSession(ctx, testutil.OwnerID, 4197, 3001, 75.00)
	if err != nil {
		t.Fatalf("OpenSession: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_sessions WHERE id = ?`, id)
	})

	if err := repo.CloseSession(ctx, testutil.OwnerID, id, 75.00, 75.00, 0.00, "test close"); err != nil {
		t.Fatalf("CloseSession: %v", err)
	}

	sess, err := repo.GetSessionByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetSessionByID after close: %v", err)
	}
	if sess.Status != domain.SessionStatusClosed {
		t.Errorf("expected status='closed', got %q", sess.Status)
	}
}

func TestUpdateSessionRefundTotals_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.OpenSession(ctx, testutil.OwnerID, 4196, 3001, 0.00)
	if err != nil {
		t.Fatalf("OpenSession: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_sessions WHERE id = ?`, id)
	})

	if err := repo.UpdateSessionRefundTotals(ctx, id, 25.00); err != nil {
		t.Fatalf("UpdateSessionRefundTotals: %v", err)
	}
}

func TestCreateOrder_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	productID := insertTestProduct(ctx, t, db, "OrderPositive", 12.50)

	orderNum := fmt.Sprintf("TEST-ORD-%d", 9001)
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
	if id <= 0 {
		t.Errorf("expected positive order id, got %d", id)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_order_items WHERE order_id = ?`, id)
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_orders WHERE id = ?`, id)
	})
}

func TestCreateOrder_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	productID := insertTestProduct(ctx, t, db, "OrderList", 15.00)

	orderNum := fmt.Sprintf("TEST-ORD-%d", 9002)
	id, err := repo.CreateOrder(ctx, testutil.OwnerID, 4101, orderNum, domain.OrderInput{
		SessionID: 3201,
		Type:      domain.OrderTypeTakeaway,
		Items: []domain.OrderItemInput{
			{MenuItemID: productID, Quantity: 2, UnitPrice: 0},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_order_items WHERE order_id = ?`, id)
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_orders WHERE id = ?`, id)
	})

	orders, _, _, err := repo.ListOrders(ctx, testutil.OwnerID, 3201, "", "", "", "", 500, 0)
	if err != nil {
		t.Fatalf("ListOrders: %v", err)
	}
	found := false
	for _, o := range orders {
		if o.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected order id=%d to appear in list", id)
	}
}

func TestUpdateOrderStatus_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	productID := insertTestProduct(ctx, t, db, "OrderStatus", 8.00)

	orderNum := fmt.Sprintf("TEST-ORD-%d", 9003)
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

	if err := repo.UpdateOrderStatus(ctx, testutil.OwnerID, id, domain.OrderStatusConfirmed); err != nil {
		t.Fatalf("UpdateOrderStatus: %v", err)
	}

	order, err := repo.GetOrder(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if order.Status != domain.OrderStatusConfirmed {
		t.Errorf("expected status='confirmed', got %q", order.Status)
	}
}

func TestGetOrder_OwnerIsolation(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetOrder(ctx, 9999999, 3301)
	if err == nil {
		t.Error("expected error when fetching order for wrong owner, got nil")
	}
}

func TestListOrders_FilterByStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
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
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.GetOrderItems(ctx, 3301)
	if err != nil {
		t.Fatalf("GetOrderItems: %v", err)
	}
	_ = items
}

func TestUpdateOrderItemStatus_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
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
	repo := infrastructure.NewRepository(db)
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
	repo := infrastructure.NewRepository(db)
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
	repo := infrastructure.NewRepository(db)
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

func TestUpdateOrderStatus_UnknownOrderReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// A stale or foreign order id used to answer success because the UPDATE
	// matched zero rows and the repo discarded the result.
	err := repo.UpdateOrderStatus(ctx, testutil.OwnerID, 99999999, domain.OrderStatusConfirmed)
	if !errors.Is(err, domain.ErrOrderNotFound) {
		t.Fatalf("UpdateOrderStatus for unknown order = %v, want ErrOrderNotFound", err)
	}
}

func TestUpdateOrderItemStatus_UnknownItemReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.UpdateOrderItemStatus(ctx, 99999999, 99999999, domain.ItemStatusPreparing)
	if !errors.Is(err, domain.ErrOrderNotFound) {
		t.Fatalf("UpdateOrderItemStatus for unknown item = %v, want ErrOrderNotFound", err)
	}
}

// MySQL reports 0 changed rows for a same-value UPDATE too, so the existence
// check must keep an idempotent re-send from looking like a missing order.
func TestUpdateOrderStatus_SameStatusIsNotReportedAsMissing(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	orderID := createSimpleTestOrder(ctx, t, db, repo, "StatusIdempotent", 1)

	order, err := repo.GetOrder(ctx, testutil.OwnerID, orderID)
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if err := repo.UpdateOrderStatus(ctx, testutil.OwnerID, orderID, order.Status); err != nil {
		t.Fatalf("same-status UpdateOrderStatus = %v, want nil", err)
	}
}

func TestUpdateOrderItemStatus_SameStatusIsNotReportedAsMissing(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	orderID := createSimpleTestOrder(ctx, t, db, repo, "ItemStatusIdempotent", 1)
	items, err := repo.GetOrderItems(ctx, orderID)
	if err != nil || len(items) != 1 {
		t.Fatalf("GetOrderItems = (%d items, %v), want 1 item", len(items), err)
	}

	if err := repo.UpdateOrderItemStatus(ctx, orderID, items[0].ID, items[0].Status); err != nil {
		t.Fatalf("same-status UpdateOrderItemStatus = %v, want nil", err)
	}
}

// createSimpleTestOrder creates one product-backed order for the fixture
// business and registers cleanup for both rows.
func createSimpleTestOrder(ctx context.Context, t *testing.T, db *sql.DB, repo *infrastructure.Repository, name string, quantity int) int64 {
	t.Helper()
	productID := insertTestProduct(ctx, t, db, name, 10.00)
	orderID, err := repo.CreateOrder(ctx, testutil.OwnerID, 4101, fmt.Sprintf("TEST-ORD-%s-%d", name, time.Now().UnixNano()), domain.OrderInput{
		SessionID: 3201,
		Type:      domain.OrderTypeTakeaway,
		Items: []domain.OrderItemInput{
			{MenuItemID: productID, Quantity: float64(quantity), UnitPrice: 10},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_order_items WHERE order_id = ?`, orderID)
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_orders WHERE id = ?`, orderID)
	})
	return orderID
}

// ─── Regression: TransferTable must not accept a foreign table ───────────────

func TestTransferTable_RejectsTableFromAnotherBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	productID := insertTestProduct(ctx, t, db, "TransferForeignTable", 10.00)

	ownTableID, err := repo.CreateTable(ctx, testutil.OwnerID, domain.TableInput{
		LocationID: 3001,
		Name:       "Transfer Own Table",
		Capacity:   2,
	})
	if err != nil {
		t.Fatalf("CreateTable own: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTable(ctx, testutil.OwnerID, ownTableID) })

	foreignTableID, err := repo.CreateTable(ctx, foreignBusinessID, domain.TableInput{
		LocationID: 3001,
		Name:       "Transfer Foreign Table",
		Capacity:   2,
	})
	if err != nil {
		t.Fatalf("CreateTable foreign: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTable(ctx, foreignBusinessID, foreignTableID) })

	orderID, err := repo.CreateOrder(ctx, testutil.OwnerID, 4101, fmt.Sprintf("TEST-ORD-TRANSFER-%d", time.Now().UnixNano()), domain.OrderInput{
		SessionID: 3201,
		TableID:   ownTableID,
		Type:      domain.OrderTypeDineIn,
		Items: []domain.OrderItemInput{
			{MenuItemID: productID, Quantity: 1, UnitPrice: 10},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_order_items WHERE order_id = ?`, orderID)
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_orders WHERE id = ?`, orderID)
	})

	err = repo.TransferTable(ctx, testutil.OwnerID, orderID, ownTableID, foreignTableID)
	if !errors.Is(err, domain.ErrTableNotFound) {
		t.Fatalf("TransferTable to foreign table = %v, want ErrTableNotFound", err)
	}

	order, err := repo.GetOrder(ctx, testutil.OwnerID, orderID)
	if err != nil {
		t.Fatalf("GetOrder after rejected transfer: %v", err)
	}
	if order.TableID != ownTableID {
		t.Errorf("order table_id = %d after rejected transfer, want %d", order.TableID, ownTableID)
	}

	foreignStatus, foreignOccupant := tableStatusAndOccupant(ctx, t, db, foreignTableID)
	if foreignStatus != "available" || foreignOccupant.Valid {
		t.Errorf("foreign table = (%q, %v), want untouched available/NULL", foreignStatus, foreignOccupant)
	}

	ownStatus, ownOccupant := tableStatusAndOccupant(ctx, t, db, ownTableID)
	if ownStatus != "occupied" || !ownOccupant.Valid || ownOccupant.Int64 != orderID {
		t.Errorf("own table = (%q, %v), want occupied by order %d after rollback", ownStatus, ownOccupant, orderID)
	}
}

// ─── Regression: payment must not free a table outside the business ─────────

func TestProcessPayment_DoesNotFreeForeignTable(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	productID := insertTestProduct(ctx, t, db, "PaymentForeignTable", 10.00)
	if _, err := db.ExecContext(ctx, `UPDATE erp_products SET track_inventory = FALSE WHERE id = ?`, productID); err != nil {
		t.Fatalf("disable inventory tracking: %v", err)
	}

	foreignTableID, err := repo.CreateTable(ctx, foreignBusinessID, domain.TableInput{
		LocationID: 3001,
		Name:       "Payment Foreign Table",
		Capacity:   2,
	})
	if err != nil {
		t.Fatalf("CreateTable foreign: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTable(ctx, foreignBusinessID, foreignTableID) })

	const foreignOccupantID int64 = 424242
	if _, err := db.ExecContext(ctx,
		`UPDATE pos_tables SET status = 'occupied', current_session_order_id = ? WHERE id = ?`,
		foreignOccupantID, foreignTableID); err != nil {
		t.Fatalf("occupy foreign table: %v", err)
	}

	orderID, err := repo.CreateOrder(ctx, testutil.OwnerID, 4101, fmt.Sprintf("TEST-ORD-PAYFOREIGN-%d", time.Now().UnixNano()), domain.OrderInput{
		LocationID: 3001,
		Type:       domain.OrderTypeDineIn,
		Items: []domain.OrderItemInput{
			{MenuItemID: productID, Quantity: 1, UnitPrice: 10},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_order_items WHERE order_id = ?`, orderID)
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_orders WHERE id = ?`, orderID)
	})
	// Simulate the legacy bad row the buggy TransferTable used to persist: an
	// order of this business pointing at another business's table.
	if _, err := db.ExecContext(ctx, `UPDATE pos_orders SET table_id = ? WHERE id = ?`, foreignTableID, orderID); err != nil {
		t.Fatalf("point order at foreign table: %v", err)
	}

	if err := repo.ProcessPayment(ctx, testutil.OwnerID, orderID, domain.PaymentInput{
		OrderID:       orderID,
		PaymentMethod: domain.PaymentMethodCash,
		AmountPaid:    10,
	}); err != nil {
		t.Fatalf("ProcessPayment: %v", err)
	}

	status, occupant := tableStatusAndOccupant(ctx, t, db, foreignTableID)
	if status != "occupied" || !occupant.Valid || occupant.Int64 != foreignOccupantID {
		t.Fatalf("foreign table after payment = (%q, %v), want occupied by %d", status, occupant, foreignOccupantID)
	}
}

func tableStatusAndOccupant(ctx context.Context, t *testing.T, db *sql.DB, tableID int64) (string, sql.NullInt64) {
	t.Helper()
	var status string
	var occupant sql.NullInt64
	if err := db.QueryRowContext(ctx,
		`SELECT status, current_session_order_id FROM pos_tables WHERE id = ?`, tableID).Scan(&status, &occupant); err != nil {
		t.Fatalf("read table %d: %v", tableID, err)
	}
	return status, occupant
}

// ─── Regression: voiding an item that is not voidable must not report success

func TestVoidOrderItem_RejectsNonVoidableItem(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	productID := insertTestProduct(ctx, t, db, "VoidNonVoidable", 10.00)

	orderID, err := repo.CreateOrder(ctx, testutil.OwnerID, 4101, fmt.Sprintf("TEST-ORD-VOIDITEM-%d", time.Now().UnixNano()), domain.OrderInput{
		SessionID: 3201,
		Type:      domain.OrderTypeTakeaway,
		Items: []domain.OrderItemInput{
			{MenuItemID: productID, Quantity: 1, UnitPrice: 10},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_order_items WHERE order_id = ?`, orderID)
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_orders WHERE id = ?`, orderID)
	})

	items, err := repo.GetOrderItems(ctx, orderID)
	if err != nil || len(items) != 1 {
		t.Fatalf("GetOrderItems = (%d items, %v), want 1 item", len(items), err)
	}
	itemID := items[0].ID
	// 'ready' items are outside the query's WHERE status IN ('pending','preparing').
	if _, err := db.ExecContext(ctx, `UPDATE pos_order_items SET status = 'ready' WHERE id = ?`, itemID); err != nil {
		t.Fatalf("mark item ready: %v", err)
	}

	err = repo.VoidOrderItem(ctx, testutil.OwnerID, orderID, itemID)
	if !errors.Is(err, domain.ErrItemNotVoidable) {
		t.Fatalf("VoidOrderItem for ready item = %v, want ErrItemNotVoidable", err)
	}

	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM pos_order_items WHERE id = ?`, itemID).Scan(&status); err != nil {
		t.Fatalf("read item status: %v", err)
	}
	if status != "ready" {
		t.Errorf("item status = %q after rejected void, want ready", status)
	}
}

// ─── Regression: splitting a paid order must be rejected ─────────────────────

func TestSplitOrder_RejectsPaidOrder(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	productID := insertTestProduct(ctx, t, db, "SplitPaidOrder", 10.00)

	orderID, err := repo.CreateOrder(ctx, testutil.OwnerID, 4101, fmt.Sprintf("TEST-ORD-SPLITPAID-%d", time.Now().UnixNano()), domain.OrderInput{
		SessionID: 3201,
		Type:      domain.OrderTypeTakeaway,
		Items: []domain.OrderItemInput{
			{MenuItemID: productID, Quantity: 1, UnitPrice: 10},
			{MenuItemID: productID, Quantity: 1, UnitPrice: 20},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_order_items WHERE order_id = ?`, orderID)
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_orders WHERE id = ?`, orderID)
	})

	items, err := repo.GetOrderItems(ctx, orderID)
	if err != nil || len(items) != 2 {
		t.Fatalf("GetOrderItems = (%d items, %v), want 2 items", len(items), err)
	}
	if _, err := db.ExecContext(ctx,
		`UPDATE pos_orders SET status = 'paid', payment_status = 'paid' WHERE id = ?`, orderID); err != nil {
		t.Fatalf("mark order paid: %v", err)
	}

	_, err = repo.SplitOrder(ctx, testutil.OwnerID, orderID, [][]int64{{items[0].ID}, {items[1].ID}})
	if !errors.Is(err, domain.ErrOrderCannotBeSplit) {
		t.Fatalf("SplitOrder for paid order = %v, want ErrOrderCannotBeSplit", err)
	}

	order, err := repo.GetOrder(ctx, testutil.OwnerID, orderID)
	if err != nil {
		t.Fatalf("GetOrder after rejected split: %v", err)
	}
	if order.Status != domain.OrderStatusPaid || order.PaymentStatus != domain.PaymentStatusPaid {
		t.Errorf("order after rejected split = (%q, %q), want paid/paid", order.Status, order.PaymentStatus)
	}
}

// ─── Regression: the session lock must not leak onto the pooled connection ──

func TestOpenSession_ReleasesAdvisoryLock(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const staffID int64 = 4195
	id, err := repo.OpenSession(ctx, testutil.OwnerID, staffID, 3001, 100.00)
	if err != nil {
		t.Fatalf("OpenSession: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_sessions WHERE id = ?`, id)
	})

	assertAdvisoryLockReleased(t, db, fmt.Sprintf("pos_session:%d:%d", testutil.OwnerID, staffID))
}

// ─── Regression: date_to must include the whole end calendar day ────────────

func TestListSessions_DateToIncludesWholeEndDay(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const staffID int64 = 4194
	id, err := repo.OpenSession(ctx, testutil.OwnerID, staffID, 3001, 0.00)
	if err != nil {
		t.Fatalf("OpenSession: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_sessions WHERE id = ?`, id)
	})

	// Opened mid-afternoon: the old LtOrEq against '2026-03-05' compared to
	// 2026-03-05 00:00:00 and dropped this session from the end day.
	if _, err := db.ExecContext(ctx,
		`UPDATE pos_sessions SET opened_at = '2026-03-05 15:30:00' WHERE id = ?`, id); err != nil {
		t.Fatalf("set opened_at: %v", err)
	}

	sessions, err := repo.ListSessions(ctx, testutil.OwnerID, 0, "2026-03-05", "2026-03-05")
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	found := false
	for _, s := range sessions {
		if s.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("session opened at 15:30 on the date_to day was omitted from ListSessions")
	}
}

func assertAdvisoryLockReleased(t *testing.T, db *sql.DB, lockName string) {
	t.Helper()
	var holder sql.NullInt64
	if err := db.QueryRowContext(context.Background(), "SELECT IS_USED_LOCK(?)", lockName).Scan(&holder); err != nil {
		t.Fatalf("check advisory lock %q: %v", lockName, err)
	}
	if holder.Valid {
		t.Fatalf("advisory lock %q is still held by connection %d", lockName, holder.Int64)
	}
}

// ─── Regression: failed CreateOrder must roll its transaction back ────────────

// CreateOrder's deferred rollback used to key off the function-scope err, which
// the table-occupancy branches shadowed, so rejecting a foreign table id left
// the transaction (and its inserted-order row locks) open. With a context that
// is never canceled, database/sql's rollback-on-cancel cannot paper over it:
// the connection must come back to the pool as soon as CreateOrder returns.
func TestCreateOrder_ForeignTableRollsBackTransaction(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	// WithoutCancel has no Done channel, so tx.awaitDone cannot roll back.
	ctx := context.WithoutCancel(context.Background())
	productID := insertTestProduct(ctx, t, db, "OrderForeignTable", 9.00)

	_, err := repo.CreateOrder(ctx, testutil.OwnerID, 4101, fmt.Sprintf("TEST-ORD-FOREIGN-%d", time.Now().UnixNano()), domain.OrderInput{
		Type:    domain.OrderTypeDineIn,
		TableID: 9999999, // no such table in this business
		Items: []domain.OrderItemInput{
			{MenuItemID: productID, Quantity: 1, UnitPrice: 0},
		},
	})
	if !errors.Is(err, domain.ErrTableNotFound) {
		t.Fatalf("CreateOrder error = %v, want ErrTableNotFound", err)
	}
	if inUse := db.Stats().InUse; inUse != 0 {
		t.Fatalf("%d connection(s) still in use after the failed CreateOrder; the transaction was not rolled back", inUse)
	}
}

// ─── Regression: close must persist the totals in the same UPDATE ─────────────

// CloseSession used to commit the close first and write the totals in a second
// statement, so a totals failure reported an error for an already-closed
// session with stale counters and no retry path. Both must now land together.
func TestCloseSessionForStaff_WritesTotalsAtomically(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const staffID int64 = 4193
	id, err := repo.OpenSession(ctx, testutil.OwnerID, staffID, 3001, 100.00)
	if err != nil {
		t.Fatalf("OpenSession: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_sessions WHERE id = ?`, id)
	})

	if err := repo.CloseSessionForStaff(ctx, testutil.OwnerID, staffID, id, 175.00, 170.00, 5.00, "atomic close",
		200.00, 100.00, 80.00, 20.00, 5); err != nil {
		t.Fatalf("CloseSessionForStaff: %v", err)
	}

	var status string
	var totalSales, cashSales, cardSales, onlineSales float64
	var totalOrders int
	if err := db.QueryRowContext(ctx, `
		SELECT status, total_sales, total_cash_sales, total_card_sales, total_online_sales, total_orders
		FROM pos_sessions WHERE id = ?`, id).
		Scan(&status, &totalSales, &cashSales, &cardSales, &onlineSales, &totalOrders); err != nil {
		t.Fatalf("read closed session: %v", err)
	}
	if status != string(domain.SessionStatusClosed) {
		t.Errorf("status = %q, want %q", status, domain.SessionStatusClosed)
	}
	if totalSales != 200.00 || cashSales != 100.00 || cardSales != 80.00 || onlineSales != 20.00 || totalOrders != 5 {
		t.Errorf("closed totals = (%v, %v, %v, %v, %d), want (200, 100, 80, 20, 5)",
			totalSales, cashSales, cardSales, onlineSales, totalOrders)
	}
}
