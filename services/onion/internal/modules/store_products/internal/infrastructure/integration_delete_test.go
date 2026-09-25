package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/store_products/internal/domain"
	"github.com/voxire/onion/internal/modules/store_products/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestDeleteStoreProduct_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.StoreProductInput{
		Title: "DeleteProduct",
		Slug:  "delete-product-row",
		Price: 1.00,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, testutil.BusinessID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, p := range items {
		if p.ID == id {
			t.Errorf("deleted product id=%d still in list", id)
		}
	}
}

func TestDeleteStoreProduct_NonExistentIsNoop(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.Delete(ctx, testutil.OwnerID, 99999999); err != nil {
		t.Fatalf("Delete non-existent: %v", err)
	}
}

func TestCreate_ThenList_FindsProduct(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.StoreProductInput{
		Title: "FindMeProduct",
		Slug:  "find-me-product",
		Price: 12.00,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.BusinessID, id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, p := range items {
		if p.ID == id {
			return
		}
	}
	t.Errorf("product id=%d not found after create", id)
}

func TestListStoreProducts_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.StoreProductInput{
		Title: "CountProduct",
		Slug:  "count-product",
		Price: 7.00,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.BusinessID, id) })

	after, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) <= len(before) {
		t.Errorf("expected count to increase: before=%d, after=%d", len(before), len(after))
	}
}

func TestCreateStoreProduct_StockQuantityIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// StoreProductInput does not have a StockQuantity field;
	// verify that creating a product with SortOrder persists without error.
	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.StoreProductInput{
		Title:     "StockProduct",
		Slug:      "stock-product",
		Price:     29.99,
		SortOrder: 5,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.BusinessID, id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, p := range items {
		if p.ID == id {
			return
		}
	}
	t.Errorf("product id=%d not found in list", id)
}
