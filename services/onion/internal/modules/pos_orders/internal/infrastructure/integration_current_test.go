package infrastructure_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

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
