package stock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/testutil"
)

func TestMovementBusinessIDDefaultsToOwnerID(t *testing.T) {
	if got := movementBusinessID(Movement{OwnerID: 42}); got != 42 {
		t.Fatalf("movementBusinessID() = %d, want owner id 42", got)
	}
}

func TestMovementBusinessIDUsesExplicitBusinessID(t *testing.T) {
	if got := movementBusinessID(Movement{OwnerID: 42, BusinessID: 99}); got != 99 {
		t.Fatalf("movementBusinessID() = %d, want explicit business id 99", got)
	}
}

func TestSetTotalTxRejectsNegativeTarget(t *testing.T) {
	_, _, err := SetTotalTx(context.Background(), nil, Movement{LocationID: 1}, -1)
	if !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("SetTotalTx() error = %v, want ErrInsufficientStock", err)
	}
}

func TestSetTotalTxRequiresLocation(t *testing.T) {
	_, _, err := SetTotalTx(context.Background(), nil, Movement{OwnerID: 42}, 1)
	if err == nil || err.Error() != "stock_core: location is required for total adjustment" {
		t.Fatalf("SetTotalTx() error = %v, want missing-location error", err)
	}
}

func TestFIFOLayerConsumption(t *testing.T) {
	got := fifoLayerConsumption([]float64{2, 5, 10}, 6)
	want := []float64{2, 4, 0}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("consumption[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestApplyMovementTxAllowsDistinctProductsForOneReference(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	ownerID := time.Now().UnixNano()
	referenceID := ownerID

	productIDs := make([]int64, 0, 2)
	for i := 0; i < 2; i++ {
		result, err := db.ExecContext(ctx,
			`INSERT INTO erp_products (owner_id, business_id, name) VALUES (?, ?, ?)`,
			ownerID, ownerID, "stock-ledger-key-test",
		)
		if err != nil {
			t.Fatalf("insert product %d: %v", i, err)
		}
		productID, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("read product %d id: %v", i, err)
		}
		productIDs = append(productIDs, productID)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM stock_ledger WHERE owner_id = ?`, ownerID)
		_, _ = db.ExecContext(ctx, `DELETE FROM stock_levels WHERE owner_id = ?`, ownerID)
		_, _ = db.ExecContext(ctx, `DELETE FROM stock_locations WHERE owner_id = ?`, ownerID)
		_, _ = db.ExecContext(ctx, `DELETE FROM erp_products WHERE owner_id = ?`, ownerID)
	})

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin stock movement transaction: %v", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	locationID, err := DefaultLocationID(ctx, tx, ownerID)
	if err != nil {
		t.Fatalf("create default location: %v", err)
	}
	for _, productID := range productIDs {
		_, _, err = ApplyMovementTx(ctx, tx, Movement{
			OwnerID:       ownerID,
			BusinessID:    ownerID,
			LocationID:    locationID,
			ProductID:     productID,
			EntryType:     EntryPurchase,
			Change:        1,
			ReferenceID:   &referenceID,
			ReferenceType: "purchase_order",
		})
		if err != nil {
			t.Fatalf("apply movement for product %d: %v", productID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit stock movements: %v", err)
	}
	committed = true

	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM stock_ledger
		WHERE owner_id = ? AND reference_type = 'purchase_order' AND reference_id = ? AND entry_type = 'purchase'`,
		ownerID, referenceID,
	).Scan(&count); err != nil {
		t.Fatalf("count stock ledger rows: %v", err)
	}
	if count != 2 {
		t.Fatalf("stock ledger contains %d rows for two products, want 2", count)
	}
}
