package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/products/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListProducts_ReturnsAllFixtureProducts(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	products, _, err := repo.ListProducts(ctx, testutil.OwnerID, "", 0, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(products) == 0 {
		t.Fatal("expected fixture rows, got none")
	}

	names := make(map[string]bool, len(products))
	for _, p := range products {
		names[p.Name] = true
	}

	for _, want := range []string{"Burger", "Soda", "Salad"} {
		if !names[want] {
			t.Errorf("expected product %q in results, got names: %v", want, names)
		}
	}
}

func TestListProducts_SaladHasLowStock(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	products, _, err := repo.ListProducts(ctx, testutil.OwnerID, "", 0, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, p := range products {
		if p.ID == 5203 {
			const wantStock = 4.0
			if p.StockQuantity != wantStock {
				t.Errorf("Salad (id=5203): expected stock_quantity=%.0f, got %.0f", wantStock, p.StockQuantity)
			}
			return
		}
	}
	t.Fatal("product id=5203 (Salad) not found in fixture rows")
}

func TestListProducts_FilterByCategoryReturnsMains(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// category_id=5002 is 'Mains'; Burger (id=5201) should be in Mains
	products, _, err := repo.ListProducts(ctx, testutil.OwnerID, "", 5002, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(products) == 0 {
		t.Fatal("expected at least one Mains product, got none")
	}

	for _, p := range products {
		if p.CategoryID != 5002 {
			t.Errorf("product %q (id=%d) has category_id=%d, want 5002", p.Name, p.ID, p.CategoryID)
		}
	}
}
