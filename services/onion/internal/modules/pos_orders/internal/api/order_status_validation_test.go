package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/testutil"
	posordersitemsv1 "github.com/voxire/voxie/gen/go/pos/orders/items/v1"
	posordersv1 "github.com/voxire/voxie/gen/go/pos/orders/v1"
)

// The seeded order lives in its own business so the fixture tenant is never
// mutated; business 9 / user 7 matches posOrdersUser.
const (
	orderStatusTestBusinessID int64 = 9
	orderStatusTestUserID     int64 = 7
)

// seedOrderStatusTest creates a business, a product, and one order with one
// item, returning the API surface plus the ids the tests probe.
func seedOrderStatusTest(t *testing.T, db *sql.DB) (context.Context, *OrderAPI, *ItemAPI, int64, int64) {
	t.Helper()
	ctx := auth.WithAuthUser(context.Background(), posOrdersUser("pos:orders:create"))
	testutil.GrantBusinessAccess(t, db, orderStatusTestBusinessID, orderStatusTestUserID)

	product, err := db.ExecContext(ctx, `
		INSERT INTO erp_products (owner_id, business_id, name, sku, sale_price, is_active, track_inventory)
		VALUES (?, ?, 'Status Test Product', 'STATUS-TEST', 5.00, TRUE, FALSE)`,
		orderStatusTestBusinessID, orderStatusTestBusinessID)
	if err != nil {
		t.Fatalf("insert test product: %v", err)
	}
	productID, err := product.LastInsertId()
	if err != nil {
		t.Fatalf("test product id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM erp_products WHERE id = ?`, productID)
	})

	service := application.NewService(ctx, db)
	orderID, err := service.CreateOrder(ctx, orderStatusTestBusinessID, orderStatusTestBusinessID, orderStatusTestUserID, domain.OrderInput{
		Type: domain.OrderTypeTakeaway,
		Items: []domain.OrderItemInput{
			{MenuItemID: productID, Quantity: 1, UnitPrice: 0},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_order_items WHERE order_id = ?`, orderID)
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_orders WHERE id = ?`, orderID)
	})

	order, err := service.GetOrder(ctx, orderStatusTestBusinessID, orderID)
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if len(order.Items) == 0 {
		t.Fatal("seeded order has no items")
	}

	return ctx, NewOrderAPI(service), NewItemAPI(service), orderID, order.Items[0].ID
}

func TestOrderAPI_UpdateOrderStatus_MapsNotFoundAndInvalidStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx, orderAPI, _, orderID, _ := seedOrderStatusTest(t, db)

	// A stale/foreign order id must not answer 200 "Order status updated".
	_, err := orderAPI.UpdateOrderStatus(ctx, &posordersv1.UpdateOrderStatusRequest{
		OrderId: 99999999,
		Body:    &posordersv1.UpdateOrderStatusBody{Status: string(domain.OrderStatusConfirmed)},
	})
	if got := posOrdersStatus(err); got != http.StatusNotFound {
		t.Fatalf("status update for unknown order = %d, want 404 (err=%v)", got, err)
	}

	// Values outside the status column's ENUM must be a 400, not a 500.
	_, err = orderAPI.UpdateOrderStatus(ctx, &posordersv1.UpdateOrderStatusRequest{
		OrderId: orderID,
		Body:    &posordersv1.UpdateOrderStatusBody{Status: "in_progress"},
	})
	if got := posOrdersStatus(err); got != http.StatusBadRequest {
		t.Fatalf("status update with unknown status = %d, want 400 (err=%v)", got, err)
	}

	// paid/refunded/voided carry companion writes and are not settable here.
	for _, status := range []domain.OrderStatus{
		domain.OrderStatusPaid,
		domain.OrderStatusRefunded,
		domain.OrderStatusVoided,
	} {
		_, err = orderAPI.UpdateOrderStatus(ctx, &posordersv1.UpdateOrderStatusRequest{
			OrderId: orderID,
			Body:    &posordersv1.UpdateOrderStatusBody{Status: string(status)},
		})
		if got := posOrdersStatus(err); got != http.StatusConflict {
			t.Fatalf("status update to %q = %d, want 409 (err=%v)", status, got, err)
		}
	}

	// A legitimate transition still succeeds.
	if _, err = orderAPI.UpdateOrderStatus(ctx, &posordersv1.UpdateOrderStatusRequest{
		OrderId: orderID,
		Body:    &posordersv1.UpdateOrderStatusBody{Status: string(domain.OrderStatusConfirmed)},
	}); err != nil {
		t.Fatalf("status update to confirmed: %v", err)
	}
}

func TestItemAPI_UpdateOrderItemStatus_MapsNotFoundAndInvalidStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx, _, itemAPI, orderID, itemID := seedOrderStatusTest(t, db)

	_, err := itemAPI.UpdateOrderItemStatus(ctx, &posordersitemsv1.UpdateOrderItemStatusRequest{
		OrderId: orderID,
		ItemId:  99999999,
		Body:    &posordersitemsv1.UpdateOrderItemStatusBody{Status: string(domain.ItemStatusPreparing)},
	})
	if got := posOrdersStatus(err); got != http.StatusNotFound {
		t.Fatalf("item status update for unknown item = %d, want 404 (err=%v)", got, err)
	}

	_, err = itemAPI.UpdateOrderItemStatus(ctx, &posordersitemsv1.UpdateOrderItemStatusRequest{
		OrderId: orderID,
		ItemId:  itemID,
		Body:    &posordersitemsv1.UpdateOrderItemStatusBody{Status: "in_progress"},
	})
	if got := posOrdersStatus(err); got != http.StatusBadRequest {
		t.Fatalf("item status update with unknown status = %d, want 400 (err=%v)", got, err)
	}

	if _, err = itemAPI.UpdateOrderItemStatus(ctx, &posordersitemsv1.UpdateOrderItemStatusRequest{
		OrderId: orderID,
		ItemId:  itemID,
		Body:    &posordersitemsv1.UpdateOrderItemStatusBody{Status: string(domain.ItemStatusPreparing)},
	}); err != nil {
		t.Fatalf("item status update to preparing: %v", err)
	}
}

// Guards the service-level validation the API relies on.
func TestService_UpdateOrderStatus_RejectsInvalidStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	service := application.NewService(context.Background(), db)
	ctx := context.Background()

	if err := service.UpdateOrderStatus(ctx, 1, 1, "in_progress"); !errors.Is(err, domain.ErrInvalidOrderStatus) {
		t.Fatalf("UpdateOrderStatus(in_progress) = %v, want ErrInvalidOrderStatus", err)
	}
	if err := service.UpdateOrderStatus(ctx, 1, 1, domain.OrderStatusPaid); !errors.Is(err, domain.ErrOrderStatusRequiresWorkflow) {
		t.Fatalf("UpdateOrderStatus(paid) = %v, want ErrOrderStatusRequiresWorkflow", err)
	}
	if err := service.UpdateOrderItemStatus(ctx, 1, 1, 1, "completed"); !errors.Is(err, domain.ErrInvalidItemStatus) {
		t.Fatalf("UpdateOrderItemStatus(completed) = %v, want ErrInvalidItemStatus", err)
	}
}
