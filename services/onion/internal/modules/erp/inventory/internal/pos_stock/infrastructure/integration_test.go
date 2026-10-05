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
