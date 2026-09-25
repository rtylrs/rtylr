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
