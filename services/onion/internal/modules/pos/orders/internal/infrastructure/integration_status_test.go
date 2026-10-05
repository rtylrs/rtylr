package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos/orders/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

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
	repo := configuredOrderRepository(db)
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
	repo := configuredOrderRepository(db)
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
	repo := configuredOrderRepository(db)
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
	repo := configuredOrderRepository(db)
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
	repo := configuredOrderRepository(db)
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
	repo := configuredOrderRepository(db)
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
