package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	onekithttp "github.com/corezio/onekit/http"

	"github.com/voxire/onion/internal/modules/inventory/internal/application"
	"github.com/voxire/onion/internal/modules/inventory/internal/domain"
	"github.com/voxire/onion/internal/modules/inventory/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	erpinventoryv1 "github.com/voxire/voxie/gen/go/erp/inventory/v1"
)

func TestAdjustInventorySaleDecrementsStock(t *testing.T) {
	db := testutil.OpenDB(t)
	const (
		businessID = int64(995901)
		userID     = int64(7)
	)

	productID := insertInventoryRegressionProduct(t, db, businessID, businessID, "inventory-sale-sign")
	ctx := inventoryRegressionContext(t, db, businessID, userID)
	registerInventoryProductCleanup(t, db, businessID, productID)

	repo := infrastructure.NewRepository(db)
	if err := repo.AdjustInventory(ctx, businessID, productID, 10, domain.InventoryTypePurchase, "seed stock"); err != nil {
		t.Fatalf("seed stock: %v", err)
	}
	before := inventoryRegressionDecimal(t, db, "SELECT stock_quantity FROM erp_products WHERE id = ?", productID)

	api := NewInventoryAPI(application.NewService(ctx, db))
	_, err := api.AdjustInventory(ctx, &erpinventoryv1.AdjustInventoryRequest{
		ProductId: testutil.Int64Ptr(productID),
		Change:    testutil.Float64Ptr(3),
		Type:      string(domain.InventoryTypeSale),
	})
	if err != nil {
		t.Fatalf("AdjustInventory(sale): %v", err)
	}

	after := inventoryRegressionDecimal(t, db, "SELECT stock_quantity FROM erp_products WHERE id = ?", productID)
	if after != before-3 {
		t.Fatalf("stock after sale = %v, want %v", after, before-3)
	}
}

func TestGetInventoryReportHonorsLimitOffsetAndReturnsTotals(t *testing.T) {
	db := testutil.OpenDB(t)
	const (
		businessID = int64(995903)
		userID     = int64(7)
	)

	productIDs := make([]int64, 0, 3)
	for i, stock := range []int{1, 2, 3} {
		productID := insertInventoryRegressionProduct(t, db, businessID, businessID, fmt.Sprintf("pagination-%d", i))
		if _, err := db.ExecContext(context.Background(), "UPDATE erp_products SET stock_quantity = ? WHERE id = ?", stock, productID); err != nil {
			t.Fatalf("set product stock: %v", err)
		}
		productIDs = append(productIDs, productID)
		registerInventoryProductCleanup(t, db, businessID, productID)
	}

	ctx := inventoryRegressionContext(t, db, businessID, userID)
	api := NewInventoryAPI(application.NewService(ctx, db))
	resp, err := api.GetInventoryReport(ctx, &erpinventoryv1.GetInventoryReportRequest{Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("GetInventoryReport: %v", err)
	}
	if got := len(resp.GetItems()); got != 1 {
		t.Fatalf("returned %d items, want 1", got)
	}
	if resp.GetLimit() != 1 || resp.GetOffset() != 1 {
		t.Fatalf("response bounds = limit %d offset %d, want 1/1", resp.GetLimit(), resp.GetOffset())
	}
	if resp.GetTotal() != int64(len(productIDs)) {
		t.Fatalf("total = %d, want %d", resp.GetTotal(), len(productIDs))
	}
	if resp.GetTotalValue() != 60 {
		t.Fatalf("total_value = %v, want 60", resp.GetTotalValue())
	}
	if !resp.GetHasMore() {
		t.Fatal("has_more = false, want true for the second page")
	}
}

func TestAdjustInventoryWasteDecrementsStock(t *testing.T) {
	db := testutil.OpenDB(t)
	const (
		businessID = int64(995904)
		userID     = int64(7)
	)

	productID := insertInventoryRegressionProduct(t, db, businessID, businessID, "inventory-waste-sign")
	ctx := inventoryRegressionContext(t, db, businessID, userID)
	registerInventoryProductCleanup(t, db, businessID, productID)

	repo := infrastructure.NewRepository(db)
	if err := repo.AdjustInventory(ctx, businessID, productID, 10, domain.InventoryTypePurchase, "seed stock"); err != nil {
		t.Fatalf("seed stock: %v", err)
	}
	before := inventoryRegressionDecimal(t, db, "SELECT stock_quantity FROM erp_products WHERE id = ?", productID)

	api := NewInventoryAPI(application.NewService(ctx, db))
	if _, err := api.AdjustInventory(ctx, &erpinventoryv1.AdjustInventoryRequest{
		ProductId: testutil.Int64Ptr(productID),
		Change:    testutil.Float64Ptr(3),
		Type:      string(domain.InventoryTypeWaste),
	}); err != nil {
		t.Fatalf("AdjustInventory(waste): %v", err)
	}

	after := inventoryRegressionDecimal(t, db, "SELECT stock_quantity FROM erp_products WHERE id = ?", productID)
	if after != before-3 {
		t.Fatalf("stock after waste = %v, want %v", after, before-3)
	}
}

func TestAdjustInventoryRejectsForeignProduct(t *testing.T) {
	db := testutil.OpenDB(t)
	const (
		callerBusinessID = int64(995902)
		userID           = int64(7)
	)

	productID := insertInventoryRegressionProduct(t, db, testutil.OtherOwnerID, testutil.OtherBusinessID, "inventory-foreign-product")
	ctx := inventoryRegressionContext(t, db, callerBusinessID, userID)
	registerInventoryProductCleanup(t, db, callerBusinessID, productID)

	api := NewInventoryAPI(application.NewService(ctx, db))
	_, err := api.AdjustInventory(ctx, &erpinventoryv1.AdjustInventoryRequest{
		ProductId: testutil.Int64Ptr(productID),
		Change:    testutil.Float64Ptr(3),
		Type:      string(domain.InventoryTypeAdjustment),
	})
	if got := inventoryRegressionErrorStatus(err); got != http.StatusNotFound {
		t.Fatalf("status = %d, want %d (err = %v)", got, http.StatusNotFound, err)
	}

	for table, query := range map[string]string{
		"stock levels":  "SELECT COUNT(*) FROM stock_levels WHERE owner_id = ? AND product_id = ?",
		"stock ledger":  "SELECT COUNT(*) FROM stock_ledger WHERE owner_id = ? AND product_id = ?",
		"legacy ledger": "SELECT COUNT(*) FROM erp_inventory_transactions WHERE owner_id = ? AND product_id = ?",
	} {
		var count int
		if err := db.QueryRowContext(ctx, query, callerBusinessID, productID).Scan(&count); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if count != 0 {
			t.Errorf("%s rows = %d, want 0", table, count)
		}
	}
}

func insertInventoryRegressionProduct(t *testing.T, db *sql.DB, ownerID, businessID int64, suffix string) int64 {
	t.Helper()
	result, err := db.ExecContext(context.Background(), `
		INSERT INTO erp_products (owner_id, business_id, name, sku, cost_price, sale_price, stock_quantity)
		VALUES (?, ?, ?, ?, 10, 20, 0)`, ownerID, businessID, "Inventory regression "+suffix, "REG-"+suffix)
	if err != nil {
		t.Fatalf("insert product: %v", err)
	}
	productID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read product id: %v", err)
	}
	return productID
}

func registerInventoryProductCleanup(t *testing.T, db *sql.DB, callerBusinessID, productID int64) {
	t.Helper()
	t.Cleanup(func() {
		ctx := context.Background()
		for _, query := range []string{
			"DELETE FROM inventory_cost_layers WHERE product_id = ?",
			"DELETE FROM stock_ledger WHERE product_id = ?",
			"DELETE FROM stock_levels WHERE product_id = ?",
			"DELETE FROM erp_inventory_transactions WHERE product_id = ?",
		} {
			if _, err := db.ExecContext(ctx, query, productID); err != nil {
				t.Errorf("cleanup %q: %v", query, err)
			}
		}
		if _, err := db.ExecContext(ctx, "DELETE FROM erp_products WHERE id = ?", productID); err != nil {
			t.Errorf("cleanup product: %v", err)
		}
		if _, err := db.ExecContext(ctx, "DELETE FROM stock_locations WHERE owner_id = ?", callerBusinessID); err != nil {
			t.Errorf("cleanup stock locations: %v", err)
		}
	})
}

func inventoryRegressionContext(t *testing.T, db *sql.DB, businessID, userID int64) context.Context {
	t.Helper()
	testutil.GrantBusinessAccess(t, db, businessID, userID)
	user := &authPkg.AuthUser{
		BusinessID: businessID,
		UserID:     userID,
		EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
			AppID:       authorization.AppERP,
			Licensed:    true,
			HasAccess:   true,
			Permissions: []string{"erp:inventory:adjust", "erp:inventory:view"},
		}}},
	}
	return authPkg.WithAuthUser(context.Background(), user)
}

func inventoryRegressionDecimal(t *testing.T, db *sql.DB, query string, args ...any) float64 {
	t.Helper()
	var value string
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&value); err != nil {
		t.Fatalf("query decimal: %v", err)
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		t.Fatalf("parse decimal %q: %v", value, err)
	}
	return parsed
}

func inventoryRegressionErrorStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}
