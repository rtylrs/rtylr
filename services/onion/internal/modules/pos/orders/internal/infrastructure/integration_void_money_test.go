package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos/orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/orders/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// Void, refund, and payment used to key their guard off one of the order's two
// state columns and write the other, so a settled sale could be relabelled
// voided and then paid out again. These cover the transitions the repositories
// now refuse, and that a refused call leaves the row byte-for-byte unchanged.

func newVoidTestOrder(ctx context.Context, t *testing.T, repo *infrastructure.Repository, db *sql.DB, label string, unitPrice float64) int64 {
	t.Helper()
	productID := insertTestProduct(ctx, t, db, label, unitPrice)

	orderID, err := repo.CreateOrder(ctx, testutil.OwnerID, 4101,
		fmt.Sprintf("TEST-ORD-%d", time.Now().UnixNano()), domain.OrderInput{
			SessionID: 3201,
			Type:      domain.OrderTypeTakeaway,
			Items: []domain.OrderItemInput{
				{MenuItemID: productID, Quantity: 1, UnitPrice: unitPrice},
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
	return orderID
}

// orderRowState is the whole money record of an order, so a refused void can be
// shown to have changed none of it.
type orderRowState struct {
	status        string
	paymentStatus string
	amountPaid    string
	notes         string
}

func readOrderRowState(ctx context.Context, t *testing.T, db *sql.DB, orderID int64) orderRowState {
	t.Helper()
	var s orderRowState
	if err := db.QueryRowContext(ctx, `
		SELECT status, COALESCE(payment_status, ''), COALESCE(CAST(amount_paid AS CHAR), ''), COALESCE(notes, '')
		FROM pos_orders WHERE id = ?`, orderID).
		Scan(&s.status, &s.paymentStatus, &s.amountPaid, &s.notes); err != nil {
		t.Fatalf("read order %d: %v", orderID, err)
	}
	return s
}

func setOrderState(ctx context.Context, t *testing.T, db *sql.DB, orderID int64, status, paymentStatus string) {
	t.Helper()
	if _, err := db.ExecContext(ctx,
		`UPDATE pos_orders SET status = ?, payment_status = ? WHERE id = ?`,
		status, paymentStatus, orderID); err != nil {
		t.Fatalf("set order %s/%s: %v", status, paymentStatus, err)
	}
}

func countOrderNotes(t *testing.T, db *sql.DB, orderID int64, marker string) int {
	t.Helper()
	var notes sql.NullString
	if err := db.QueryRowContext(context.Background(),
		`SELECT notes FROM pos_orders WHERE id = ?`, orderID).Scan(&notes); err != nil {
		t.Fatalf("read notes for order %d: %v", orderID, err)
	}
	return strings.Count(notes.String, marker)
}

func TestVoidOrder_VoidsUnpaidOrder(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()
	orderID := newVoidTestOrder(ctx, t, repo, db, "VoidUnpaid", 10.00)

	if err := repo.VoidOrder(ctx, testutil.OwnerID, orderID, "customer left"); err != nil {
		t.Fatalf("VoidOrder: %v", err)
	}

	got := readOrderRowState(ctx, t, db, orderID)
	if got.status != string(domain.OrderStatusVoided) {
		t.Errorf("status = %q after void, want voided", got.status)
	}
	if got.paymentStatus != string(domain.PaymentStatusUnpaid) {
		t.Errorf("payment_status = %q after void, want unpaid", got.paymentStatus)
	}
	if n := countOrderNotes(t, db, orderID, "[Voided: customer left]"); n != 1 {
		t.Errorf("void reason recorded %d time(s), want 1", n)
	}
}

func TestVoidOrder_IsIdempotentAndRecordsOneReason(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()
	orderID := newVoidTestOrder(ctx, t, repo, db, "VoidTwice", 12.00)

	if err := repo.VoidOrder(ctx, testutil.OwnerID, orderID, "first reason"); err != nil {
		t.Fatalf("first VoidOrder: %v", err)
	}
	// A re-sent void is a duplicate request, not a new business event: it must
	// succeed without writing a second reason into the audit trail.
	if err := repo.VoidOrder(ctx, testutil.OwnerID, orderID, "second reason"); err != nil {
		t.Fatalf("second VoidOrder: %v", err)
	}

	if n := countOrderNotes(t, db, orderID, "[Voided: "); n != 1 {
		t.Fatalf("void reasons recorded %d time(s), want 1", n)
	}
	if n := countOrderNotes(t, db, orderID, "second reason"); n != 0 {
		t.Fatalf("re-sent void appended its reason %d time(s), want 0", n)
	}
}

func TestVoidOrder_RejectsPaidOrderWithoutSideEffects(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()
	orderID := newVoidTestOrder(ctx, t, repo, db, "VoidPaid", 40.00)
	if _, err := db.ExecContext(ctx, `
		UPDATE pos_orders
		SET status = 'paid', payment_status = 'paid', payment_method = 'cash', amount_paid = 40.00, paid_at = NOW()
		WHERE id = ?`, orderID); err != nil {
		t.Fatalf("mark order paid: %v", err)
	}
	before := readOrderRowState(ctx, t, db, orderID)

	if err := repo.VoidOrder(ctx, testutil.OwnerID, orderID, "manager override"); !errors.Is(err, domain.ErrOrderAlreadyPaid) {
		t.Fatalf("VoidOrder for paid order = %v, want ErrOrderAlreadyPaid", err)
	}

	if after := readOrderRowState(ctx, t, db, orderID); after != before {
		t.Errorf("paid order changed by the rejected void:\n before %+v\n after  %+v", before, after)
	}
}

func TestVoidOrder_RejectsRefundedOrderWithoutSideEffects(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()
	orderID := newVoidTestOrder(ctx, t, repo, db, "VoidRefunded", 25.00)
	setOrderState(ctx, t, db, orderID, string(domain.OrderStatusRefunded), string(domain.PaymentStatusRefunded))
	before := readOrderRowState(ctx, t, db, orderID)

	if err := repo.VoidOrder(ctx, testutil.OwnerID, orderID, "manager override"); !errors.Is(err, domain.ErrOrderAlreadyPaid) {
		t.Fatalf("VoidOrder for refunded order = %v, want ErrOrderAlreadyPaid", err)
	}

	if after := readOrderRowState(ctx, t, db, orderID); after != before {
		t.Errorf("refunded order changed by the rejected void:\n before %+v\n after  %+v", before, after)
	}
}

func TestVoidOrder_RejectsCancelledOrder(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()
	orderID := newVoidTestOrder(ctx, t, repo, db, "VoidCancelled", 15.00)
	setOrderState(ctx, t, db, orderID, string(domain.OrderStatusCancelled), string(domain.PaymentStatusUnpaid))
	before := readOrderRowState(ctx, t, db, orderID)

	if err := repo.VoidOrder(ctx, testutil.OwnerID, orderID, "manager override"); !errors.Is(err, domain.ErrOrderNotEditable) {
		t.Fatalf("VoidOrder for cancelled order = %v, want ErrOrderNotEditable", err)
	}

	if after := readOrderRowState(ctx, t, db, orderID); after != before {
		t.Errorf("cancelled order changed by the rejected void:\n before %+v\n after  %+v", before, after)
	}
}

func TestVoidOrder_ReportsUnknownOrder(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)

	if err := repo.VoidOrder(context.Background(), testutil.OwnerID, 999999999, "no such order"); !errors.Is(err, domain.ErrOrderNotFound) {
		t.Fatalf("VoidOrder for unknown order = %v, want ErrOrderNotFound", err)
	}
}

// A void used to leave payment_status at 'paid', so orders voided before the
// guard existed must still refuse a refund: paying one out credits the
// merchant for a sale the record says never happened.
func TestCreateRefund_RejectsVoidedOrderThatTookMoney(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()
	orderID := newVoidTestOrder(ctx, t, repo, db, "RefundVoidedPaid", 33.00)
	if _, err := db.ExecContext(ctx, `
		UPDATE pos_orders
		SET status = 'voided', payment_status = 'paid', amount_paid = 33.00, notes = ' [Voided: legacy]'
		WHERE id = ?`, orderID); err != nil {
		t.Fatalf("build legacy voided order: %v", err)
	}

	_, err := repo.CreateRefund(ctx, testutil.OwnerID, orderID, 4101, 33.00, "duplicate credit")
	if !errors.Is(err, domain.ErrOrderNotEditable) {
		t.Fatalf("CreateRefund for voided order = %v, want ErrOrderNotEditable", err)
	}

	var refunds int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM pos_refunds WHERE order_id = ?`, orderID).Scan(&refunds); err != nil {
		t.Fatalf("count refunds: %v", err)
	}
	if refunds != 0 {
		t.Fatalf("%d refund row(s) written for a rejected refund, want 0", refunds)
	}
}

func TestProcessPayment_RejectsVoidedOrder(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()
	orderID := newVoidTestOrder(ctx, t, repo, db, "PayVoided", 18.00)

	if err := repo.VoidOrder(ctx, testutil.OwnerID, orderID, "customer left"); err != nil {
		t.Fatalf("VoidOrder: %v", err)
	}
	before := readOrderRowState(ctx, t, db, orderID)

	err := repo.ProcessPayment(ctx, testutil.OwnerID, orderID, domain.PaymentInput{
		OrderID:       orderID,
		PaymentMethod: domain.PaymentMethodCash,
		AmountPaid:    18.00,
	})
	if !errors.Is(err, domain.ErrOrderNotEditable) {
		t.Fatalf("ProcessPayment for voided order = %v, want ErrOrderNotEditable", err)
	}

	if after := readOrderRowState(ctx, t, db, orderID); after != before {
		t.Errorf("voided order was resurrected by the rejected payment:\n before %+v\n after  %+v", before, after)
	}
}

func TestProcessPayment_RejectsRefundedOrder(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()
	orderID := newVoidTestOrder(ctx, t, repo, db, "PayRefunded", 27.00)
	setOrderState(ctx, t, db, orderID, string(domain.OrderStatusRefunded), string(domain.PaymentStatusRefunded))
	before := readOrderRowState(ctx, t, db, orderID)

	err := repo.ProcessPayment(ctx, testutil.OwnerID, orderID, domain.PaymentInput{
		OrderID:       orderID,
		PaymentMethod: domain.PaymentMethodCard,
		AmountPaid:    27.00,
	})
	if !errors.Is(err, domain.ErrOrderNotEditable) {
		t.Fatalf("ProcessPayment for refunded order = %v, want ErrOrderNotEditable", err)
	}

	if after := readOrderRowState(ctx, t, db, orderID); after != before {
		t.Errorf("refunded order took a second payment:\n before %+v\n after  %+v", before, after)
	}
}
