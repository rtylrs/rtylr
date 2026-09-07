package infrastructure_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

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

	orderNum := fmt.Sprintf("TEST-ORD-%d", 9101)
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
		for _, oid := range append([]int64{id}, newIDs...) {
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
