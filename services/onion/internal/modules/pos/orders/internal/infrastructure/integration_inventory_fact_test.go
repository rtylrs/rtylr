package infrastructure_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	poscontract "github.com/voxire/onion/internal/contracts/pos"
	"github.com/voxire/onion/internal/modules/erp/inventory"
	"github.com/voxire/onion/internal/modules/pos"
	"github.com/voxire/onion/internal/modules/pos/locations"
	"github.com/voxire/onion/internal/modules/pos/orders/internal/domain"
	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/testutil"
)

func TestPaidOrderFactUpdatesERPInventoryOnce(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := configuredOrderRepository(db)
	productID := insertTestProduct(ctx, t, db, fmt.Sprintf("paid-fact-%d", time.Now().UnixNano()), 10)
	if _, err := db.ExecContext(ctx, "UPDATE erp_products SET track_inventory = TRUE WHERE id = ?", productID); err != nil {
		t.Fatal(err)
	}
	orderID, err := repo.CreateOrder(ctx, testutil.BusinessID, 4101, fmt.Sprintf("PAID-%d", time.Now().UnixNano()), domain.OrderInput{
		SessionID: 3201, LocationID: 3001, Type: domain.OrderTypeTakeaway,
		Items: []domain.OrderItemInput{
			{MenuItemID: productID, Name: "line one", Quantity: 2, UnitPrice: 10},
			{MenuItemID: productID, Name: "line two", Quantity: 3, UnitPrice: 10},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM pos_outbox WHERE subject_type IN ('order', 'pos_order') AND subject_id = ?", orderID)
		_, _ = db.ExecContext(ctx, "DELETE FROM platform_events WHERE subject_type IN ('order', 'pos_order') AND subject_id = ?", orderID)
		_, _ = db.ExecContext(ctx, "DELETE FROM stock_ledger WHERE product_id = ?", productID)
		_, _ = db.ExecContext(ctx, "DELETE FROM stock_levels WHERE product_id = ?", productID)
		_, _ = db.ExecContext(ctx, "DELETE FROM pos_order_items WHERE order_id = ?", orderID)
		_, _ = db.ExecContext(ctx, "DELETE FROM pos_orders WHERE id = ?", orderID)
	})
	stock := func() float64 {
		t.Helper()
		var quantity float64
		if err := db.QueryRowContext(ctx, "SELECT COALESCE(SUM(on_hand), 0) FROM stock_levels WHERE business_id = ? AND product_id = ?", testutil.BusinessID, productID).Scan(&quantity); err != nil {
			t.Fatal(err)
		}
		return quantity
	}
	if err := repo.ProcessPayment(ctx, testutil.BusinessID, orderID, domain.PaymentInput{AmountPaid: 50, PaymentMethod: domain.PaymentMethodCash}); err != nil {
		t.Fatal(err)
	}
	if stock() != 0 {
		t.Fatal("POS transaction mutated ERP stock")
	}
	consumer := inventory.NewModule(ctx, db).WithBusinessOwnerResolver(func(_ context.Context, businessID int64) (int64, error) {
		if businessID != testutil.BusinessID {
			return 0, errors.New("foreign business")
		}
		return testutil.OwnerID, nil
	}).WithPOSBranchOwnerReader(locations.NewModule(ctx, db).StockBranchOwner)
	dispatcher := eventbus.NewDispatcher(db, consumer).WithSources(pos.Facts(db))
	if _, err := dispatcher.DispatchBatch(ctx); err != nil {
		t.Fatal(err)
	}
	var fact eventbus.Event
	fact.Type, fact.SubjectType, fact.SubjectID, fact.BusinessID = poscontract.OrderPaidType, "pos_order", orderID, testutil.BusinessID
	var raw []byte
	if err := db.QueryRowContext(ctx, "SELECT id, payload FROM platform_events WHERE business_id = ? AND event_type = ? AND subject_type = ? AND subject_id = ?", fact.BusinessID, fact.Type, fact.SubjectType, orderID).Scan(&fact.ID, &raw); err != nil {
		t.Fatal(err)
	}
	fact.Payload = raw
	for replay := 0; replay < 2; replay++ {
		if err := consumer.DispatchEvent(ctx, fact); err != nil {
			t.Fatal(err)
		}
		if quantity := stock(); quantity != -5 {
			t.Fatalf("inventory after replay %d = %v, want -5", replay, quantity)
		}
	}
	var ledgerCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM stock_ledger WHERE business_id = ? AND product_id = ? AND reference_type = 'pos_order' AND reference_id = ?", testutil.BusinessID, productID, orderID).Scan(&ledgerCount); err != nil {
		t.Fatal(err)
	}
	if ledgerCount != 1 {
		t.Fatalf("ledger entries = %d, want one", ledgerCount)
	}
	if err := repo.ProcessPayment(ctx, testutil.BusinessID, orderID, domain.PaymentInput{AmountPaid: 50, PaymentMethod: domain.PaymentMethodCash}); !errors.Is(err, domain.ErrOrderAlreadyPaid) {
		t.Fatalf("duplicate payment = %v", err)
	}
}
