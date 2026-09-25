package infrastructure_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const foreignBusinessID int64 = 9002

func insertTestProduct(ctx context.Context, t *testing.T, db *sql.DB, name string, salePrice float64) int64 {
	t.Helper()
	result, err := db.ExecContext(ctx, `
		INSERT INTO erp_products (owner_id, business_id, name, sku, sale_price, is_active)
		VALUES (?, ?, ?, ?, ?, TRUE)
	`, testutil.OwnerID, testutil.OwnerID, name, "TEST-"+name, salePrice)
	if err != nil {
		t.Fatalf("insert test product: %v", err)
	}
	productID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("test product id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM erp_products WHERE id = ?`, productID)
	})
	return productID
}

func TestListLocations_ReturnsFixtureRows(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	locations, err := repo.ListLocations(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locations) == 0 {
		t.Fatal("expected fixture rows, got none")
	}

	var found bool
	for _, loc := range locations {
		if loc.ID == 3001 && loc.Name == "Test Branch" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected location id=3001 named 'Test Branch', got %+v", locations)
	}
}

func TestGetOrder_ReturnsFixturePaidOrder(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	order, err := repo.GetOrder(ctx, testutil.OwnerID, 3301)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order == nil {
		t.Fatal("expected order, got nil")
	}
	if string(order.Status) != "paid" {
		t.Errorf("expected status='paid', got %q", order.Status)
	}
	const wantSubtotal = 150.00
	if order.Subtotal != wantSubtotal {
		t.Errorf("expected subtotal=%.2f, got %.2f", wantSubtotal, order.Subtotal)
	}
}

func TestListOrders_ReturnsBothFixtureOrders(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// sessionID=3201 has 2 orders in fixtures
	orders, _, _, err := repo.ListOrders(ctx, testutil.OwnerID, 3201, "", "", "", "", 500, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orders) < 2 {
		t.Fatalf("expected at least 2 orders, got %d", len(orders))
	}
	foundWithItems := false
	for _, order := range orders {
		if order.ID == 3301 && len(order.Items) > 0 {
			foundWithItems = true
			break
		}
	}
	if !foundWithItems {
		t.Fatalf("expected listed fixture order 3301 to include purchased items, got %+v", orders)
	}
}

func TestCreateLocation_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateLocation(ctx, testutil.OwnerID, domain.LocationInput{
		Name:    "Test Create Location",
		Address: "123 Main St",
		Phone:   "555-0001",
	})
	if err != nil {
		t.Fatalf("CreateLocation: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive id, got %d", id)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_locations WHERE id = ?`, id)
	})
}

func TestCreateLocation_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateLocation(ctx, testutil.OwnerID, domain.LocationInput{
		Name: "Appear In List Location",
	})
	if err != nil {
		t.Fatalf("CreateLocation: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_locations WHERE id = ?`, id)
	})

	locations, err := repo.ListLocations(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListLocations: %v", err)
	}
	found := false
	for _, loc := range locations {
		if loc.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected location id=%d to appear in list", id)
	}
}

func TestListLocations_IsolatedByOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	locations, err := repo.ListLocations(ctx, 9999999)
	if err != nil {
		t.Fatalf("ListLocations: %v", err)
	}
	for _, loc := range locations {
		if loc.OwnerID == testutil.OwnerID {
			t.Errorf("isolation violated: location id=%d belongs to fixture owner", loc.ID)
		}
	}
}

func TestCreateTable_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateTable(ctx, testutil.OwnerID, domain.TableInput{
		LocationID: 3001,
		Name:       "Table T1",
		Capacity:   4,
	})
	if err != nil {
		t.Fatalf("CreateTable: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive id, got %d", id)
	}
	t.Cleanup(func() {
		_ = repo.DeleteTable(ctx, testutil.OwnerID, id)
	})
}

func TestListTables_ForLocation(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateTable(ctx, testutil.OwnerID, domain.TableInput{
		LocationID: 3001,
		Name:       "Filter Location Table",
		Capacity:   2,
	})
	if err != nil {
		t.Fatalf("CreateTable: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.DeleteTable(ctx, testutil.OwnerID, id)
	})

	tables, err := repo.ListTables(ctx, testutil.OwnerID, 3001, "")
	if err != nil {
		t.Fatalf("ListTables: %v", err)
	}
	found := false
	for _, tbl := range tables {
		if tbl.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected table id=%d to appear in list for locationID=3001", id)
	}
}
