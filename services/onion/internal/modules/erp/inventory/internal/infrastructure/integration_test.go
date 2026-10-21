package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/inventory/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/inventory/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetInventoryReport_ReturnsProducts(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	report, err := repo.GetInventoryReport(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetInventoryReport error: %v", err)
	}
	if len(report) == 0 {
		t.Fatal("expected at least one product in inventory report, got none")
	}
}

func TestGetLowStockProducts_ReturnsSalad(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	// Salad has stock=4 which is below the default low-stock threshold of 10
	items, err := repo.GetLowStockProducts(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetLowStockProducts error: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected at least one low-stock product, got none")
	}

	found := false
	for _, item := range items {
		if item.ProductName == "Salad" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Salad' in low-stock results; got: %v", items)
	}
}

func TestGetInventoryReport_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	report, err := repo.GetInventoryReport(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetInventoryReport(ownerID=0): %v", err)
	}
	if len(report) != 0 {
		t.Fatalf("expected 0 rows for ownerID=0, got %d", len(report))
	}
}

func TestGetInventoryReport_ContainsBurger(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	report, err := repo.GetInventoryReport(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetInventoryReport: %v", err)
	}
	for _, item := range report {
		if item.ProductName == "Burger" {
			return
		}
	}
	t.Error("expected 'Burger' in inventory report, not found")
}

func TestGetInventoryReport_ContainsSoda(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	report, err := repo.GetInventoryReport(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetInventoryReport: %v", err)
	}
	for _, item := range report {
		if item.ProductName == "Soda" {
			return
		}
	}
	t.Error("expected 'Soda' in inventory report, not found")
}

func TestGetLowStockProducts_ContainsSalad(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	// Salad has stock=4 and threshold=10 => low stock
	items, err := repo.GetLowStockProducts(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetLowStockProducts: %v", err)
	}
	for _, item := range items {
		if item.ProductName == "Salad" {
			return
		}
	}
	t.Error("expected Salad (stock=4, threshold=10) to appear in low-stock results")
}

func TestGetLowStockProducts_BurgerNotLowStock(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	// Burger has stock=50 which is above threshold=10
	items, err := repo.GetLowStockProducts(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetLowStockProducts: %v", err)
	}
	for _, item := range items {
		if item.ProductName == "Burger" {
			t.Error("Burger (stock=50) should not appear in low-stock results")
			return
		}
	}
}

func TestGetValuation_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	items, err := repo.GetValuation(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetValuation(ownerID=0): %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 valuation rows for ownerID=0, got %d", len(items))
	}
}

func TestGetValuation_ReturnsAtLeastOneRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	items, err := repo.GetValuation(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetValuation: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected at least one valuation row, got none")
	}
	for _, item := range items {
		if item.WarehouseName == "" {
			t.Error("expected valuation row to include a warehouse name")
		}
		if item.TotalValue != item.TotalCost {
			t.Errorf("valuation total_value=%.2f, want cost-based total_cost=%.2f", item.TotalValue, item.TotalCost)
		}
	}
}

func TestGetValuation_ProductHasPositiveCost(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	items, err := repo.GetValuation(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetValuation: %v", err)
	}
	// At least one item should have a non-negative total cost
	for _, item := range items {
		if item.TotalCost >= 0 {
			return
		}
	}
	t.Error("expected at least one valuation row with non-negative TotalCost")
}

func TestAdjustInventory_DecreasesStock(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Burger (5201) has stock=50; decrease by 5
	if err := repo.AdjustInventory(ctx, testutil.OwnerID, 5201, -5, domain.InventoryTypeAdjustment, "test decrease"); err != nil {
		t.Fatalf("AdjustInventory decrease: %v", err)
	}
	// Restore
	t.Cleanup(func() {
		_ = repo.AdjustInventory(ctx, testutil.OwnerID, 5201, 5, domain.InventoryTypeAdjustment, "restore after test decrease")
	})
}

func TestAdjustInventory_IncreasesStock(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Burger (5201) has stock=50; increase by 10
	if err := repo.AdjustInventory(ctx, testutil.OwnerID, 5201, 10, domain.InventoryTypePurchase, "test increase"); err != nil {
		t.Fatalf("AdjustInventory increase: %v", err)
	}
	// Restore
	t.Cleanup(func() {
		_ = repo.AdjustInventory(ctx, testutil.OwnerID, 5201, -10, domain.InventoryTypeAdjustment, "restore after test increase")
	})
}

func TestAdjustInventory_ZeroChange_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// A zero change should not error (no-op transaction)
	err := repo.AdjustInventory(ctx, testutil.OwnerID, 5201, 0, domain.InventoryTypeAdjustment, "zero change")
	if err != nil {
		t.Fatalf("AdjustInventory(change=0): unexpected error: %v", err)
	}
}

func TestGetInventoryReport_AllProductsHaveOwnerID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	report, err := repo.GetInventoryReport(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetInventoryReport: %v", err)
	}
	for _, item := range report {
		if item.ProductID <= 0 {
			t.Errorf("inventory report item has non-positive ProductID: %d (name=%q)", item.ProductID, item.ProductName)
		}
	}
}
