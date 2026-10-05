package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos/orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/orders/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// MarkOrderRefunded used to be an unconditional
// `SET status='refunded', payment_status='refunded'`, and the over-refund
// guard compared float64. A penny refund on a $100 sale therefore deleted the
// whole sale from GetSessionTotals and the dashboard while pos_refunds
// recorded a penny, and a legitimate full refund in awkward cents was refused
// with a 400 because 0.03 + 0.27 is 0.30000000000000004 in binary floating
// point. These exercise the corrected behaviour end to end.

func paidTestOrder(ctx context.Context, t *testing.T, repo *infrastructure.Repository, db *sql.DB, label string, total float64) int64 {
	t.Helper()
	orderID := newVoidTestOrder(ctx, t, repo, db, label, total)
	if _, err := db.ExecContext(ctx, `
		UPDATE pos_orders
		SET status = 'paid', payment_status = 'paid', payment_method = 'cash',
		    amount_paid = ?, paid_at = NOW()
		WHERE id = ?`, total, orderID); err != nil {
		t.Fatalf("mark order paid: %v", err)
	}
	return orderID
}

func orderMoneyState(ctx context.Context, t *testing.T, db *sql.DB, orderID int64) orderRowState {
	t.Helper()
	return readOrderRowState(ctx, t, db, orderID)
}

func refundedTotal(ctx context.Context, t *testing.T, db *sql.DB, orderID int64) string {
	t.Helper()
	var total sql.NullString
	if err := db.QueryRowContext(ctx,
		`SELECT CAST(COALESCE(SUM(amount), 0) AS CHAR) FROM pos_refunds WHERE order_id = ?`,
		orderID).Scan(&total); err != nil {
		t.Fatalf("sum refunds for %d: %v", orderID, err)
	}
	return total.String
}

// A penny against a $100 sale must leave the order in every revenue query.
func TestCreateRefund_PartialRefundKeepsTheOrderInTheBooks(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()
	orderID := paidTestOrder(ctx, t, repo, db, "RefundPartialKeeps", 100.00)

	if _, err := repo.CreateRefund(ctx, testutil.OwnerID, orderID, 4101, 0.01, "one item returned"); err != nil {
		t.Fatalf("CreateRefund of 0.01: %v", err)
	}

	got := orderMoneyState(ctx, t, db, orderID)
	if got.paymentStatus != string(domain.PaymentStatusPaid) {
		t.Errorf("payment_status = %q after a $0.01 refund on a $100 sale, want %q: the order drops out of every revenue query",
			got.paymentStatus, domain.PaymentStatusPaid)
	}
	if got.status != string(domain.OrderStatusPaid) {
		t.Errorf("status = %q after a partial refund, want %q", got.status, domain.OrderStatusPaid)
	}

	// The sale must still be counted by the session aggregate that reads
	// payment_status = 'paid'.
	var counted string
	if err := db.QueryRowContext(ctx, `
		SELECT CAST(COALESCE(SUM(total), 0) AS CHAR) FROM pos_orders
		WHERE id = ? AND payment_status = 'paid'`, orderID).Scan(&counted); err != nil {
		t.Fatalf("revenue aggregate: %v", err)
	}
	if counted != "100.00" {
		t.Errorf("revenue aggregate = %s after a partial refund, want 100.00", counted)
	}

	if sum := refundedTotal(ctx, t, db, orderID); sum != "0.01" {
		t.Errorf("pos_refunds total = %s, want 0.01", sum)
	}
}

// A full refund, in cents that do not survive float64 addition, must be
// accepted and must retire the sale.
func TestCreateRefund_FullRefundInAwkwardCentsIsAccepted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()
	orderID := paidTestOrder(ctx, t, repo, db, "RefundAwkwardCents", 0.30)

	if _, err := repo.CreateRefund(ctx, testutil.OwnerID, orderID, 4101, 0.03, "first part"); err != nil {
		t.Fatalf("CreateRefund of 0.03: %v", err)
	}
	// 0.03 + 0.27 is 0.30000000000000004 as a float64, which is greater than
	// 0.30. The pre-fix guard compared exactly that and refused this refund.
	if _, err := repo.CreateRefund(ctx, testutil.OwnerID, orderID, 4101, 0.27, "remainder"); err != nil {
		t.Fatalf("last cent of a legitimate full refund was refused: %v", err)
	}

	got := orderMoneyState(ctx, t, db, orderID)
	if got.paymentStatus != string(domain.PaymentStatusRefunded) {
		t.Errorf("payment_status = %q after a full refund, want %q", got.paymentStatus, domain.PaymentStatusRefunded)
	}
	if got.status != string(domain.OrderStatusRefunded) {
		t.Errorf("status = %q after a full refund, want %q", got.status, domain.OrderStatusRefunded)
	}
	if sum := refundedTotal(ctx, t, db, orderID); sum != "0.30" {
		t.Errorf("pos_refunds total = %s, want 0.30", sum)
	}
}

// A whole dollar refunded one cent at a time must reach the cap exactly, and
// the next cent must be refused.
func TestCreateRefund_CumulativeRefundsStopAtTheCap(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()
	orderID := paidTestOrder(ctx, t, repo, db, "RefundCumulative", 1.00)

	for i := 0; i < 100; i++ {
		if _, err := repo.CreateRefund(ctx, testutil.OwnerID, orderID, 4101, 0.01, "cent by cent"); err != nil {
			t.Fatalf("refund %d of 0.01 refused: %v", i+1, err)
		}
	}
	if _, err := repo.CreateRefund(ctx, testutil.OwnerID, orderID, 4101, 0.01, "one cent too many"); !errors.Is(err, domain.ErrRefundExceedsTotal) {
		t.Fatalf("refund past the cap = %v, want ErrRefundExceedsTotal", err)
	}

	got := orderMoneyState(ctx, t, db, orderID)
	if got.paymentStatus != string(domain.PaymentStatusRefunded) {
		t.Errorf("payment_status = %q after refunding the whole cap, want %q", got.paymentStatus, domain.PaymentStatusRefunded)
	}
	if sum := refundedTotal(ctx, t, db, orderID); sum != "1.00" {
		t.Errorf("pos_refunds total = %s, want 1.00", sum)
	}

	var rows int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM pos_refunds WHERE order_id = ?`, orderID).Scan(&rows); err != nil {
		t.Fatalf("count refunds: %v", err)
	}
	if rows != 100 {
		t.Errorf("%d refund rows written, want 100: the refused refund must not be persisted", rows)
	}
}

// Cash tender carries the change handed straight back, so the cap is the net
// the drawer kept. Capping at amount_paid would pay out change the merchant
// never had.
func TestCreateRefund_CapExcludesChangeAlreadyHandedBack(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()
	orderID := paidTestOrder(ctx, t, repo, db, "RefundChangeCap", 100.00)
	if _, err := db.ExecContext(ctx,
		`UPDATE pos_orders SET amount_paid = 120.00, change_due = 20.00 WHERE id = ?`, orderID); err != nil {
		t.Fatalf("record change: %v", err)
	}

	if _, err := repo.CreateRefund(ctx, testutil.OwnerID, orderID, 4101, 100.00, "full sale back"); err != nil {
		t.Fatalf("refunding the net kept was refused: %v", err)
	}
	if _, err := repo.CreateRefund(ctx, testutil.OwnerID, orderID, 4101, 0.01, "change was already given back"); !errors.Is(err, domain.ErrRefundExceedsTotal) {
		t.Fatalf("refund beyond the net kept = %v, want ErrRefundExceedsTotal", err)
	}
	if sum := refundedTotal(ctx, t, db, orderID); sum != "100.00" {
		t.Errorf("pos_refunds total = %s, want 100.00", sum)
	}
}

// A voided order is still refused, including the legacy shape where its
// payment_status still reads paid.
func TestCreateRefund_VoidedOrderIsStillRefusedAfterAPartialRefund(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()
	orderID := paidTestOrder(ctx, t, repo, db, "RefundVoidedAfterPartial", 60.00)

	if _, err := repo.CreateRefund(ctx, testutil.OwnerID, orderID, 4101, 10.00, "part returned"); err != nil {
		t.Fatalf("partial refund: %v", err)
	}
	if err := repo.VoidOrder(ctx, testutil.OwnerID, orderID, "mistake"); !errors.Is(err, domain.ErrOrderAlreadyPaid) {
		t.Fatalf("VoidOrder after a partial refund = %v, want ErrOrderAlreadyPaid", err)
	}
	// Model a legacy voided row; the current API correctly refused the void above.
	if _, err := db.ExecContext(ctx, "UPDATE pos_orders SET status = 'voided' WHERE id = ?", orderID); err != nil {
		t.Fatalf("seed legacy voided order: %v", err)
	}
	if _, err := repo.CreateRefund(ctx, testutil.OwnerID, orderID, 4101, 10.00, "after void"); !errors.Is(err, domain.ErrOrderNotEditable) {
		t.Fatalf("refund of a voided order = %v, want ErrOrderNotEditable", err)
	}
}

// A paid order refuses a second charge, so a partial refund cannot be followed
// by re-charging the customer.
func TestProcessPayment_RefusesSecondChargeAfterAPartialRefund(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()
	orderID := paidTestOrder(ctx, t, repo, db, "PayAfterPartial", 80.00)

	if _, err := repo.CreateRefund(ctx, testutil.OwnerID, orderID, 4101, 20.00, "part returned"); err != nil {
		t.Fatalf("partial refund: %v", err)
	}

	err := repo.ProcessPayment(ctx, testutil.OwnerID, orderID, domain.PaymentInput{
		OrderID:       orderID,
		PaymentMethod: domain.PaymentMethodCash,
		AmountPaid:    80.00,
	})
	if !errors.Is(err, domain.ErrOrderAlreadyPaid) {
		t.Fatalf("second charge after a partial refund = %v, want ErrOrderAlreadyPaid", err)
	}
	if sum := refundedTotal(ctx, t, db, orderID); sum != "20.00" {
		t.Errorf("pos_refunds total = %s, want 20.00", sum)
	}
}

// Register and public checkouts share a business-locked number allocator.
func TestCreateOrder_ConcurrentCreatesAllGetDistinctOrderNumbers(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()
	productID := insertTestProduct(ctx, t, db, "OrderConcurrentNumbers", 10.00)

	const workers = 8
	ids := make([]int64, workers)
	errs := make([]error, workers)

	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			if i%2 == 0 {
				ids[i], _, errs[i] = repo.CreatePublicOrder(ctx, testutil.BusinessID, "", "Test Customer", "", "", []domain.OrderItemInput{
					{MenuItemID: productID, Name: "Test product", Quantity: 1, UnitPrice: 10.00},
				})
				return
			}
			ids[i], errs[i] = repo.CreateOrder(ctx, testutil.OwnerID, 4101, "", domain.OrderInput{
				SessionID: 3201,
				Type:      domain.OrderTypeTakeaway,
				Items: []domain.OrderItemInput{
					{MenuItemID: productID, Quantity: 1, UnitPrice: 10.00},
				},
			})
		}(i)
	}
	close(start)
	wg.Wait()

	created := make([]int64, 0, workers)
	for i := 0; i < workers; i++ {
		if errs[i] != nil {
			t.Errorf("concurrent checkout %d failed: %v", i, errs[i])
			continue
		}
		created = append(created, ids[i])
	}
	t.Cleanup(func() {
		for _, id := range created {
			_, _ = db.ExecContext(context.Background(), `DELETE FROM pos_order_items WHERE order_id = ?`, id)
			_, _ = db.ExecContext(context.Background(), `DELETE FROM pos_orders WHERE id = ?`, id)
		}
	})

	seen := make(map[string]int64, len(created))
	for _, id := range created {
		order, err := repo.GetOrder(ctx, testutil.OwnerID, id)
		if err != nil {
			t.Fatalf("GetOrder %d: %v", id, err)
		}
		if prev, dup := seen[order.OrderNumber]; dup {
			t.Errorf("orders %d and %d share order number %q", prev, id, order.OrderNumber)
		}
		seen[order.OrderNumber] = id
	}
	if len(seen) != len(created) {
		t.Errorf("%d orders created but only %d distinct numbers", len(created), len(seen))
	}
}

// The generated number must be produced by the transaction, not handed in
// from the service on a separate connection: a caller that passes one still
// gets exactly that number.
func TestCreateOrder_HonoursAnExplicitOrderNumber(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredOrderRepository(db)
	ctx := context.Background()
	productID := insertTestProduct(ctx, t, db, "OrderExplicitNumber", 5.00)

	want := fmt.Sprintf("TEST-ORD-EXPLICIT-%d", time.Now().UnixNano())
	id, err := repo.CreateOrder(ctx, testutil.OwnerID, 4101, want, domain.OrderInput{
		SessionID: 3201,
		Type:      domain.OrderTypeTakeaway,
		Items: []domain.OrderItemInput{
			{MenuItemID: productID, Quantity: 1, UnitPrice: 5.00},
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
	if order.OrderNumber != want {
		t.Errorf("order number = %q, want the explicit %q", order.OrderNumber, want)
	}
}
