package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/inventory/internal/infrastructure"
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
