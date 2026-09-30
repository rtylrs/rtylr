package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/testutil"
	posordersv1 "github.com/voxire/voxie/gen/go/pos/orders/v1"
)

// The repository refuses a void on an order that already took money or is no
// longer editable, and answers with a typed error. Every void handler used to
// map anything it did not recognise to 500, so those refusals reached the till
// as "Failed to void order" instead of a conflict.

func voidConflictOrder(t *testing.T, db *sql.DB, label, status, paymentStatus string) (context.Context, *OrderAPI, *VoidOrderHandler, int64) {
	t.Helper()
	ctx := auth.WithAuthUser(context.Background(), posOrdersUser("pos:orders:void"))
	testutil.GrantBusinessAccess(t, db, orderStatusTestBusinessID, orderStatusTestUserID)

	product, err := db.ExecContext(ctx, `
		INSERT INTO erp_products (owner_id, business_id, name, sku, sale_price, is_active, track_inventory)
		VALUES (?, ?, 'Void Conflict Product', ?, 5.00, TRUE, FALSE)`,
		orderStatusTestBusinessID, orderStatusTestBusinessID, "VOID-CONFLICT-"+label)
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
			{MenuItemID: productID, Quantity: 1, UnitPrice: 5.00},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_refunds WHERE order_id = ?`, orderID)
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_order_items WHERE order_id = ?`, orderID)
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_orders WHERE id = ?`, orderID)
	})

	if _, err := db.ExecContext(ctx,
		`UPDATE pos_orders SET status = ?, payment_status = ? WHERE id = ?`,
		status, paymentStatus, orderID); err != nil {
		t.Fatalf("set order %s/%s: %v", status, paymentStatus, err)
	}

	return ctx, NewOrderAPI(service), &VoidOrderHandler{Service: service}, orderID
}

// The mapping itself needs no database: it is the shared helper all three void
// entry points now call.
func TestVoidOrderError_MapsTypedRefusals(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"unknown order is not found", domain.ErrOrderNotFound, http.StatusNotFound},
		{"a settled order is a conflict", domain.ErrOrderAlreadyPaid, http.StatusConflict},
		{"an uneditable order is a conflict", domain.ErrOrderNotEditable, http.StatusConflict},
		{
			// The repository wraps some of these; errors.Is must still see them.
			name: "a wrapped conflict is a conflict",
			err:  fmt.Errorf("void order: %w", domain.ErrOrderAlreadyPaid),
			want: http.StatusConflict,
		},
		{"an unexpected failure is a server error", errors.New("connection reset"), http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := posOrdersStatus(voidOrderError(tc.err)); got != tc.want {
				t.Fatalf("voidOrderError(%v) = %d, want %d", tc.err, got, tc.want)
			}
		})
	}
}

func TestVoidOrder_RefusalsAre409Not500(t *testing.T) {
	db := testutil.OpenDB(t)

	cases := []struct {
		name          string
		status        string
		paymentStatus string
	}{
		{
			name:          "a settled order cannot be voided",
			status:        string(domain.OrderStatusPaid),
			paymentStatus: string(domain.PaymentStatusPaid),
		},
		{
			name:          "a refunded order cannot be voided",
			status:        string(domain.OrderStatusRefunded),
			paymentStatus: string(domain.PaymentStatusRefunded),
		},
		{
			name:          "a cancelled order cannot be voided",
			status:        string(domain.OrderStatusCancelled),
			paymentStatus: string(domain.PaymentStatusUnpaid),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, orderAPI, handler, orderID := voidConflictOrder(t, db, tc.name, tc.status, tc.paymentStatus)

			// The gRPC-style OrderAPI void.
			_, err := orderAPI.VoidOrder(ctx, &posordersv1.VoidOrderRequest{
				OrderId: orderID,
				Reason:  "manager override",
			})
			if got := posOrdersStatus(err); got != http.StatusConflict {
				t.Errorf("OrderAPI.VoidOrder = %d, want 409 (err=%v)", got, err)
			}

			// The HTTP VoidOrderHandler void, which had the same mapping gap.
			_, err = handler.Serve(ctx, &VoidOrderRequest{OrderID: orderID, Reason: "manager override"})
			if got := posOrdersStatus(err); got != http.StatusConflict {
				t.Errorf("VoidOrderHandler.Serve = %d, want 409 (err=%v)", got, err)
			}

			// The approve/override handler, the third site.
			_, err = serveVoidOrderWithPermission(ctx, &VoidOrderApprovalRequest{OrderID: orderID, Reason: "manager override"},
				handler.Service, "pos:orders:void")
			if got := posOrdersStatus(err); got != http.StatusConflict {
				t.Errorf("serveVoidOrderWithPermission = %d, want 409 (err=%v)", got, err)
			}
		})
	}
}

func TestVoidOrder_MissingOrderIsStill404(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx, orderAPI, handler, _ := voidConflictOrder(t, db, "NotFound", string(domain.OrderStatusPending), string(domain.PaymentStatusUnpaid))

	_, err := orderAPI.VoidOrder(ctx, &posordersv1.VoidOrderRequest{
		OrderId: 99999999,
		Reason:  "typo",
	})
	if got := posOrdersStatus(err); got != http.StatusNotFound {
		t.Errorf("OrderAPI.VoidOrder for unknown order = %d, want 404 (err=%v)", got, err)
	}

	_, err = handler.Serve(ctx, &VoidOrderRequest{OrderID: 99999999, Reason: "typo"})
	if got := posOrdersStatus(err); got != http.StatusNotFound {
		t.Errorf("VoidOrderHandler.Serve for unknown order = %d, want 404 (err=%v)", got, err)
	}
}
