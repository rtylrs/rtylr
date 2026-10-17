package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/inventory/internal/pos_stock/domain"
	"github.com/voxire/onion/internal/modules/erp/inventory/internal/pos_stock/infrastructure"
	"github.com/voxire/onion/internal/modules/organization/business"
	"github.com/voxire/onion/internal/modules/pos/locations"
	"github.com/voxire/onion/internal/testutil"
)

// Fixture IDs: business_id=1001, pos_locations.id=3001, erp_products.id=5201.
const (
	fixtureBusinessID int64 = testutil.BusinessID
	fixtureOwnerID    int64 = 1001
	fixtureBranchID   int64 = 3001
	fixtureProductID  int64 = 5201
)

func TestStockMovements_ListMovements_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredStockRepository(db)

	mvts, err := repo.ListMovements(context.Background(), fixtureBusinessID, fixtureOwnerID, fixtureBranchID, fixtureProductID)
	if err != nil {
		t.Fatalf("ListMovements: %v", err)
	}
	_ = mvts
}

func TestStockMovements_ListMovements_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredStockRepository(db)

	mvts, err := repo.ListMovements(context.Background(), fixtureBusinessID, fixtureOwnerID, fixtureBranchID, fixtureProductID)
	if err != nil {
		t.Fatalf("ListMovements: %v", err)
	}
	if mvts == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
}

func TestStockMovements_ListMovements_WrongBusinessReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredStockRepository(db)

	mvts, err := repo.ListMovements(context.Background(), 999999, 999999, fixtureBranchID, fixtureProductID)
	if err != nil {
		t.Fatalf("ListMovements wrong business: %v", err)
	}
	if len(mvts) != 0 {
		t.Fatalf("expected 0 movements for unknown business, got %d", len(mvts))
	}
}

func TestStockMovements_CreateMovement_ReturnsNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredStockRepository(db)
	ctx := context.Background()

	movType := "receive"
	id, err := repo.CreateMovement(ctx, fixtureBusinessID, fixtureBranchID, fixtureProductID, movType, 10.0, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("CreateMovement: %v", err)
	}
	if id == 0 {
		t.Fatal("CreateMovement returned ID=0")
	}
}

func TestStockMovements_CreateMovement_ThenListContainsNew(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredStockRepository(db)
	ctx := context.Background()

	note := "integration list check"
	id, err := repo.CreateMovement(ctx, fixtureBusinessID, fixtureBranchID, fixtureProductID, "receive", 5.0, nil, nil, &note, nil)
	if err != nil {
		t.Fatalf("CreateMovement: %v", err)
	}

	mvts, err := repo.ListMovements(ctx, fixtureBusinessID, fixtureOwnerID, fixtureBranchID, fixtureProductID)
	if err != nil {
		t.Fatalf("ListMovements: %v", err)
	}
	found := false
	for _, m := range mvts {
		if m.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("new movement id=%d not found in list", id)
	}
}

func TestStockMovements_ListMovements_AllHaveNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredStockRepository(db)
	ctx := context.Background()

	_, _ = repo.CreateMovement(ctx, fixtureBusinessID, fixtureBranchID, fixtureProductID, "receive", 1.0, nil, nil, nil, nil)

	mvts, err := repo.ListMovements(ctx, fixtureBusinessID, fixtureOwnerID, fixtureBranchID, fixtureProductID)
	if err != nil {
		t.Fatalf("ListMovements: %v", err)
	}
	for i, m := range mvts {
		if m.ID == 0 {
			t.Errorf("mvts[%d].ID is zero", i)
		}
	}
}

func TestStockMovements_ListMovements_AllHaveNonEmptyType(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredStockRepository(db)
	ctx := context.Background()

	_, _ = repo.CreateMovement(ctx, fixtureBusinessID, fixtureBranchID, fixtureProductID, "receive", 1.0, nil, nil, nil, nil)

	mvts, err := repo.ListMovements(ctx, fixtureBusinessID, fixtureOwnerID, fixtureBranchID, fixtureProductID)
	if err != nil {
		t.Fatalf("ListMovements: %v", err)
	}
	for i, m := range mvts {
		if m.Type == "" {
			t.Errorf("mvts[%d].Type is empty", i)
		}
	}
}

func TestStockMovements_UpdateInventory_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredStockRepository(db)

	err := repo.UpdateInventory(context.Background(), fixtureBusinessID, fixtureBranchID, fixtureProductID, 5.0)
	if err != nil {
		t.Fatalf("UpdateInventory: %v", err)
	}
}

func TestStockMovements_SetInventoryQuantity_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredStockRepository(db)

	err := repo.SetInventoryQuantity(context.Background(), fixtureBusinessID, fixtureBranchID, fixtureProductID, 100.0)
	if err != nil {
		t.Fatalf("SetInventoryQuantity: %v", err)
	}
}

func TestStockMovements_GetInventoryItem_AfterSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredStockRepository(db)
	ctx := context.Background()

	err := repo.SetInventoryQuantity(ctx, fixtureBusinessID, fixtureBranchID, fixtureProductID, 77.0)
	if err != nil {
		t.Fatalf("SetInventoryQuantity: %v", err)
	}

	item, err := repo.GetInventoryItem(ctx, fixtureBusinessID, fixtureOwnerID, fixtureBranchID, fixtureProductID)
	if err != nil {
		t.Fatalf("GetInventoryItem: %v", err)
	}
	if item.Quantity != 77.0 {
		t.Errorf("Quantity = %f, want 77.0", item.Quantity)
	}
}

func TestStockMovements_GetInventoryItem_UnknownReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredStockRepository(db)

	_, err := repo.GetInventoryItem(context.Background(), 999999, 999999, 999999, 999999)
	if !errors.Is(err, domain.ErrInventoryNotFound) {
		t.Fatalf("expected ErrInventoryNotFound, got %v", err)
	}
}

func TestStockMovements_ListBranchInventory_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredStockRepository(db)

	items, err := repo.ListBranchInventory(context.Background(), fixtureBusinessID, fixtureOwnerID, fixtureBranchID)
	if err != nil {
		t.Fatalf("ListBranchInventory: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestStockMovements_ListBranchInventory_AllHaveNonNegativeQuantity(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredStockRepository(db)

	items, err := repo.ListBranchInventory(context.Background(), fixtureBusinessID, fixtureOwnerID, fixtureBranchID)
	if err != nil {
		t.Fatalf("ListBranchInventory: %v", err)
	}
	for i, item := range items {
		if item.Quantity < 0 {
			t.Errorf("items[%d].Quantity = %f, want >= 0", i, item.Quantity)
		}
	}
}

func TestStockMovements_ListLowStockItems_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredStockRepository(db)

	items, err := repo.ListLowStockItems(context.Background(), fixtureBusinessID, fixtureOwnerID, fixtureBranchID)
	if err != nil {
		t.Fatalf("ListLowStockItems: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice")
	}
}

func configuredStockRepository(db *sql.DB) *infrastructure.Repository {
	ctx := context.Background()
	return infrastructure.NewRepository(db).
		WithBusinessOwnerResolver(business.NewModule(ctx, db).GetBusinessOwnerID).
		WithBranchOwnerReader(locations.NewModule(ctx, db).StockBranchOwner)
}

func TestStockMovementsRejectForeignBranchWithSameOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	result, err := db.ExecContext(ctx, `INSERT INTO pos_locations (owner_id, business_id, name) VALUES (?, ?, 'Foreign branch')`, fixtureOwnerID, testutil.OtherBusinessID)
	if err != nil {
		t.Fatal(err)
	}
	branchID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM pos_locations WHERE id = ?", branchID) })
	repo := configuredStockRepository(db)
	var before float64
	if err := db.QueryRowContext(ctx, "SELECT stock_quantity FROM erp_products WHERE id = ?", fixtureProductID).Scan(&before); err != nil {
		t.Fatal(err)
	}
	_, err = repo.ReceiveStock(ctx, domain.ReceiveStockInput{BusinessID: fixtureBusinessID, OwnerID: fixtureOwnerID, BranchID: branchID, ProductID: fixtureProductID, Quantity: 2})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("foreign branch should be unavailable, got %v", err)
	}
	var after float64
	if err := db.QueryRowContext(ctx, "SELECT stock_quantity FROM erp_products WHERE id = ?", fixtureProductID).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatalf("rejected branch changed inventory from %v to %v", before, after)
	}
}

func TestStockMovementsUseBusinessOwnerForMemberRequest(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredStockRepository(db)
	id, err := repo.ReceiveStock(context.Background(), domain.ReceiveStockInput{BusinessID: fixtureBusinessID, OwnerID: 1002, BranchID: fixtureBranchID, ProductID: fixtureProductID, Quantity: 2})
	if err != nil {
		t.Fatal(err)
	}
	var owner, businessID int64
	if err := db.QueryRow("SELECT owner_id, business_id FROM stock_ledger WHERE id = ?", id).Scan(&owner, &businessID); err != nil {
		t.Fatal(err)
	}
	if owner != fixtureOwnerID || businessID != fixtureBusinessID {
		t.Fatalf("ledger ownership = %d/%d", owner, businessID)
	}
}

func TestStockMovements_TransferStockPairsLegsAndCarriesCost(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredStockRepository(db)
	ctx := context.Background()

	result, err := db.ExecContext(ctx, `INSERT INTO pos_locations (owner_id, business_id, name) VALUES (?, ?, 'Transfer destination')`, fixtureOwnerID, fixtureBusinessID)
	if err != nil {
		t.Fatal(err)
	}
	destinationBranchID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	result, err = db.ExecContext(ctx, `
		INSERT INTO erp_products (owner_id, business_id, name, cost_price, track_inventory)
		VALUES (?, ?, 'pos-transfer-cost', 9.00, TRUE)`, fixtureOwnerID, fixtureBusinessID)
	if err != nil {
		t.Fatal(err)
	}
	productID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, query := range []string{
			"DELETE FROM inventory_cost_layers WHERE owner_id = ? AND product_id = ?",
			"DELETE FROM stock_ledger WHERE owner_id = ? AND product_id = ?",
			"DELETE FROM stock_levels WHERE owner_id = ? AND product_id = ?",
		} {
			_, _ = db.ExecContext(ctx, query, fixtureOwnerID, productID)
		}
		_, _ = db.ExecContext(ctx, "DELETE FROM stock_locations WHERE source_table = 'pos_locations' AND source_id = ?", destinationBranchID)
		_, _ = db.ExecContext(ctx, "DELETE FROM pos_locations WHERE id = ?", destinationBranchID)
	})

	if _, err := repo.ReceiveStock(ctx, domain.ReceiveStockInput{BusinessID: fixtureBusinessID, OwnerID: fixtureOwnerID, BranchID: fixtureBranchID, ProductID: productID, Quantity: 5}); err != nil {
		t.Fatalf("ReceiveStock: %v", err)
	}
	inID, err := repo.TransferStock(ctx, domain.TransferStockInput{
		BusinessID: fixtureBusinessID, OwnerID: fixtureOwnerID,
		FromBranchID: fixtureBranchID, ToBranchID: destinationBranchID,
		ProductID: productID, Quantity: 3,
	})
	if err != nil {
		t.Fatalf("TransferStock: %v", err)
	}

	var entryType string
	var group, reference sql.NullInt64
	if err := db.QueryRowContext(ctx, "SELECT entry_type, transfer_group, reference_id FROM stock_ledger WHERE id = ?", inID).Scan(&entryType, &group, &reference); err != nil {
		t.Fatalf("read inbound leg: %v", err)
	}
	if entryType != "transfer_in" || !group.Valid || !reference.Valid || group.Int64 != reference.Int64 {
		t.Fatalf("inbound leg = %s group=%v reference=%v, want transfer_in referencing its group", entryType, group, reference)
	}
	var outType string
	var outGroup sql.NullInt64
	if err := db.QueryRowContext(ctx, "SELECT entry_type, transfer_group FROM stock_ledger WHERE id = ?", group.Int64).Scan(&outType, &outGroup); err != nil {
		t.Fatalf("read outbound leg: %v", err)
	}
	if outType != "transfer_out" || outGroup.Int64 != group.Int64 {
		t.Fatalf("outbound leg = %s group=%v, want transfer_out in group %d", outType, outGroup, group.Int64)
	}

	branchStock := func(branchID int64) (onHand, layered, value float64) {
		t.Helper()
		if err := db.QueryRowContext(ctx, `
			SELECT sl.on_hand,
			       COALESCE((SELECT SUM(quantity_remaining) FROM inventory_cost_layers c WHERE c.owner_id = sl.owner_id AND c.product_id = sl.product_id AND c.location_id = sl.location_id), 0),
			       COALESCE((SELECT SUM(quantity_remaining * unit_cost) FROM inventory_cost_layers c WHERE c.owner_id = sl.owner_id AND c.product_id = sl.product_id AND c.location_id = sl.location_id), 0)
			FROM stock_levels sl
			JOIN stock_locations l ON l.id = sl.location_id
			WHERE sl.owner_id = ? AND sl.product_id = ? AND l.source_table = 'pos_locations' AND l.source_id = ?`,
			fixtureOwnerID, productID, branchID).Scan(&onHand, &layered, &value); err != nil {
			t.Fatalf("read branch %d stock: %v", branchID, err)
		}
		return onHand, layered, value
	}
	if onHand, layered, value := branchStock(fixtureBranchID); onHand != 2 || layered != 2 || value != 18 {
		t.Errorf("source branch = %v on hand, %v layered worth %v; want 2, 2, 18", onHand, layered, value)
	}
	if onHand, layered, value := branchStock(destinationBranchID); onHand != 3 || layered != 3 || value != 27 {
		t.Errorf("destination branch = %v on hand, %v layered worth %v; want 3, 3, 27", onHand, layered, value)
	}
}
