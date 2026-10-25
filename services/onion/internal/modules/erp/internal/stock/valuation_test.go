package stock

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/testutil"
)

type valuationFixture struct {
	db        *sql.DB
	ownerID   int64
	productID int64
	warehouse int64
	branch    int64
}

func newValuationFixture(t *testing.T, costPrice float64) valuationFixture {
	t.Helper()
	db := testutil.OpenDB(t)
	ctx := context.Background()
	ownerID := time.Now().UnixNano()

	result, err := db.ExecContext(ctx,
		`INSERT INTO erp_products (owner_id, business_id, name, cost_price) VALUES (?, ?, ?, ?)`,
		ownerID, ownerID, "valuation-test", costPrice,
	)
	if err != nil {
		t.Fatalf("insert product: %v", err)
	}
	productID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read product id: %v", err)
	}
	t.Cleanup(func() {
		for _, table := range []string{"inventory_cost_layers", "stock_ledger", "stock_levels", "stock_locations", "erp_products"} {
			_, _ = db.ExecContext(ctx, "DELETE FROM "+table+" WHERE owner_id = ?", ownerID)
		}
	})

	f := valuationFixture{db: db, ownerID: ownerID, productID: productID}
	f.inTx(t, func(tx *sql.Tx) error {
		var err error
		if f.warehouse, err = DefaultLocationID(ctx, tx, ownerID); err != nil {
			return err
		}
		f.branch, err = POSBranchLocationID(ctx, tx, ownerID, 9001)
		return err
	})
	return f
}

func (f valuationFixture) inTx(t *testing.T, fn func(tx *sql.Tx) error) {
	t.Helper()
	tx, err := f.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		t.Fatalf("stock operation: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
}

func (f valuationFixture) movement(location int64, entry EntryType, change float64) Movement {
	return Movement{
		BusinessID: f.ownerID,
		LocationID: location,
		ProductID:  f.productID,
		EntryType:  entry,
		Change:     change,
	}
}

func (f valuationFixture) receive(t *testing.T, location int64, quantity, unitCost float64) {
	t.Helper()
	f.inTx(t, func(tx *sql.Tx) error {
		m := f.movement(location, EntryPurchase, quantity)
		m.UnitCost = &unitCost
		_, _, err := ApplyMovementTx(context.Background(), tx, m)
		return err
	})
}

// layers returns the location's open layers as quantity and total value.
func (f valuationFixture) layers(t *testing.T, location int64) (quantity, value float64) {
	t.Helper()
	if err := f.db.QueryRowContext(context.Background(), `
		SELECT COALESCE(SUM(quantity_remaining), 0), COALESCE(SUM(quantity_remaining * unit_cost), 0)
		FROM inventory_cost_layers
		WHERE owner_id = ? AND product_id = ? AND location_id = ? AND quantity_remaining > 0`,
		f.ownerID, f.productID, location,
	).Scan(&quantity, &value); err != nil {
		t.Fatalf("read layers: %v", err)
	}
	return quantity, value
}

func (f valuationFixture) onHand(t *testing.T, location int64) float64 {
	t.Helper()
	var onHand float64
	if err := f.db.QueryRowContext(context.Background(),
		`SELECT on_hand FROM stock_levels WHERE owner_id = ? AND product_id = ? AND location_id = ?`,
		f.ownerID, f.productID, location,
	).Scan(&onHand); err != nil {
		t.Fatalf("read on_hand: %v", err)
	}
	return onHand
}

func TestTransferTxCarriesSourceCostLayers(t *testing.T) {
	f := newValuationFixture(t, 9)
	f.receive(t, f.warehouse, 10, 2)
	f.receive(t, f.warehouse, 5, 4)

	f.inTx(t, func(tx *sql.Tx) error {
		_, _, err := TransferTx(context.Background(), tx,
			f.movement(f.warehouse, EntryTransferOut, -12),
			f.movement(f.branch, EntryTransferIn, 12))
		return err
	})

	if got := f.onHand(t, f.warehouse); got != 3 {
		t.Errorf("source on_hand = %v, want 3", got)
	}
	if got := f.onHand(t, f.branch); got != 12 {
		t.Errorf("destination on_hand = %v, want 12", got)
	}
	// FIFO: the 10 units at 2.00 leave first, then 2 of the 5 at 4.00.
	if quantity, value := f.layers(t, f.warehouse); quantity != 3 || value != 12 {
		t.Errorf("source layers = %v units worth %v, want 3 worth 12", quantity, value)
	}
	if quantity, value := f.layers(t, f.branch); quantity != 12 || value != 28 {
		t.Errorf("destination layers = %v units worth %v, want 12 worth 28", quantity, value)
	}
}

func TestTransferTxValuesUnlayeredStockAtCostPrice(t *testing.T) {
	f := newValuationFixture(t, 1.5)
	// Stock that predates cost layers: on hand, no layer.
	if _, err := f.db.ExecContext(context.Background(), `
		INSERT INTO stock_levels (owner_id, business_id, location_id, product_id, on_hand)
		VALUES (?, ?, ?, ?, 8)`, f.ownerID, f.ownerID, f.warehouse, f.productID); err != nil {
		t.Fatalf("seed legacy stock: %v", err)
	}

	f.inTx(t, func(tx *sql.Tx) error {
		_, _, err := TransferTx(context.Background(), tx,
			f.movement(f.warehouse, EntryTransferOut, -4),
			f.movement(f.branch, EntryTransferIn, 4))
		return err
	})

	if quantity, value := f.layers(t, f.branch); quantity != 4 || value != 6 {
		t.Errorf("destination layers = %v units worth %v, want 4 worth 6", quantity, value)
	}
}

func TestTransferTxRejectsMismatchedLegs(t *testing.T) {
	out := Movement{BusinessID: 1, LocationID: 1, ProductID: 1, Change: -2}
	for name, in := range map[string]Movement{
		"different product":  {BusinessID: 1, LocationID: 2, ProductID: 2, Change: 2},
		"different quantity": {BusinessID: 1, LocationID: 2, ProductID: 1, Change: 3},
		"same location":      {BusinessID: 1, LocationID: 1, ProductID: 1, Change: 2},
	} {
		if _, _, err := TransferTx(context.Background(), nil, out, in); err == nil {
			t.Errorf("%s: TransferTx accepted mismatched legs", name)
		}
	}
}

func TestInboundMovementWithoutUnitCostUsesProductCost(t *testing.T) {
	f := newValuationFixture(t, 3.5)
	f.inTx(t, func(tx *sql.Tx) error {
		_, _, err := ApplyMovementTx(context.Background(), tx, f.movement(f.warehouse, EntryAdjustment, 4))
		return err
	})
	if quantity, value := f.layers(t, f.warehouse); quantity != 4 || value != 14 {
		t.Errorf("layers = %v units worth %v, want 4 worth 14", quantity, value)
	}
}

func TestInboundMovementOnlyLayersStockAboveZero(t *testing.T) {
	f := newValuationFixture(t, 1)
	f.inTx(t, func(tx *sql.Tx) error {
		sale := f.movement(f.warehouse, EntrySale, -3)
		sale.AllowNegative = true
		_, _, err := ApplyMovementTx(context.Background(), tx, sale)
		return err
	})
	f.receive(t, f.warehouse, 5, 2)

	if got := f.onHand(t, f.warehouse); got != 2 {
		t.Fatalf("on_hand = %v, want 2", got)
	}
	if quantity, value := f.layers(t, f.warehouse); quantity != 2 || value != 4 {
		t.Errorf("layers = %v units worth %v, want 2 worth 4", quantity, value)
	}
}

func TestSetLevelTxKeepsLayersInStepWithCount(t *testing.T) {
	f := newValuationFixture(t, 2)
	f.receive(t, f.warehouse, 5, 3)

	count := func(target float64) {
		t.Helper()
		f.inTx(t, func(tx *sql.Tx) error {
			_, _, err := SetLevelTx(context.Background(), tx, f.movement(f.warehouse, EntryCount, 0), target)
			return err
		})
	}
	count(9)
	if quantity, value := f.layers(t, f.warehouse); quantity != 9 || value != 23 {
		t.Errorf("after counting a gain, layers = %v units worth %v, want 9 worth 23", quantity, value)
	}
	count(6)
	if quantity, value := f.layers(t, f.warehouse); quantity != 6 || value != 14 {
		t.Errorf("after counting a loss, layers = %v units worth %v, want 6 worth 14", quantity, value)
	}
}

func TestApplyMovementTxSkipsARepeatedReference(t *testing.T) {
	f := newValuationFixture(t, 1)
	reference := f.productID
	apply := func() {
		t.Helper()
		f.inTx(t, func(tx *sql.Tx) error {
			m := f.movement(f.warehouse, EntryReturn, 2)
			m.ReferenceID = &reference
			m.ReferenceType = "store_order"
			_, _, err := ApplyMovementTx(context.Background(), tx, m)
			return err
		})
	}
	apply()
	apply()
	if got := f.onHand(t, f.warehouse); got != 2 {
		t.Errorf("on_hand = %v, want 2 after a redelivered movement", got)
	}
	if quantity, _ := f.layers(t, f.warehouse); quantity != 2 {
		t.Errorf("layers = %v units, want 2 after a redelivered movement", quantity)
	}
}

func TestInsufficientStockOpensNoLayer(t *testing.T) {
	f := newValuationFixture(t, 1)
	tx, err := f.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	_, _, err = ApplyMovementTx(context.Background(), tx, f.movement(f.warehouse, EntryWaste, -1))
	if !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("error = %v, want ErrInsufficientStock", err)
	}
}
