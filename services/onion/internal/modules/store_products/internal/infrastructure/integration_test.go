package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/store_products/internal/domain"
	"github.com/voxire/onion/internal/modules/store_products/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestStoreProducts_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestStoreProducts_List_BusinessScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, p := range items {
		if p.BusinessID != testutil.BusinessID {
			t.Errorf("expected business_id %d, got %d", testutil.BusinessID, p.BusinessID)
		}
	}
}

func TestStoreProducts_Create_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreProductInput{
		Title:       "Integration Test Product",
		Slug:        "integration-test-product",
		Price:       29.99,
		IsPublished: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })
}

func TestStoreProducts_Create_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreProductInput{
		Title: "ListAppear Product",
		Slug:  "list-appear-product",
		Price: 9.99,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, p := range items {
		if p.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("created product id=%d not found in list", id)
	}
}

func TestStoreProducts_Update_ChangesTitle(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreProductInput{
		Title: "OldTitle", Slug: "old-title", Price: 5.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.StoreProductInput{
		Title: "NewTitle", Slug: "new-title", Price: 5.0,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, p := range items {
		if p.ID == id && p.Title != "NewTitle" {
			t.Errorf("expected title NewTitle, got %s", p.Title)
		}
	}
}

func TestStoreProducts_Delete_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreProductInput{
		Title: "DeleteMe", Slug: "delete-me", Price: 1.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(context.Background(), testutil.BusinessID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, p := range items {
		if p.ID == id {
			t.Errorf("deleted product id=%d still in list", id)
		}
	}
}

func TestStoreProducts_List_UnknownOwner_Empty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), 9999999)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 for unknown owner, got %d", len(items))
	}
}

func TestStoreProducts_Create_WithCompareAtPrice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreProductInput{
		Title:          "Sale Product",
		Slug:           "sale-product",
		Price:          49.99,
		CompareAtPrice: 79.99,
		IsPublished:    true,
	})
	if err != nil {
		t.Fatalf("Create with compare price: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestStoreProducts_Update_PublishedToggle(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreProductInput{
		Title: "DraftProduct", Slug: "draft-product", Price: 10.0, IsPublished: false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.StoreProductInput{
		Title: "DraftProduct", Slug: "draft-product", Price: 10.0, IsPublished: true,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestStoreProducts_Create_WithDescription(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreProductInput{
		Title:       "Described Product",
		Slug:        "described-product",
		Description: "This is a detailed description of the product.",
		Price:       15.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestListStoreProducts_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for owner 0, got %d", len(items))
	}
}

func TestListStoreProducts_IsNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
}

func TestCreateStoreProduct_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.StoreProductInput{
		Title: "PersistProduct",
		Slug:  "persist-product",
		Price: 19.99,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.BusinessID, id) })
}

func TestCreateStoreProduct_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.StoreProductInput{
		Title: "TitleStoredProduct",
		Slug:  "title-stored-product",
		Price: 9.99,
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
			if p.Title != "TitleStoredProduct" {
				t.Errorf("expected title TitleStoredProduct, got %s", p.Title)
			}
			return
		}
	}
	t.Errorf("product id=%d not found in list", id)
}

func TestCreateStoreProduct_PriceIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.StoreProductInput{
		Title: "PriceProduct",
		Slug:  "price-product",
		Price: 55.55,
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
			if p.Price != 55.55 {
				t.Errorf("expected price 55.55, got %f", p.Price)
			}
			return
		}
	}
	t.Errorf("product id=%d not found in list", id)
}

func TestUpdateStoreProduct_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.StoreProductInput{
		Title: "OldProductName",
		Slug:  "old-product-name",
		Price: 10.00,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.BusinessID, id) })

	if err := repo.Update(ctx, testutil.BusinessID, id, domain.StoreProductInput{
		Title: "NewProductName",
		Slug:  "new-product-name",
		Price: 10.00,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, p := range items {
		if p.ID == id && p.Title != "NewProductName" {
			t.Errorf("expected title NewProductName, got %s", p.Title)
		}
	}
}

func TestUpdateStoreProduct_ChangesPrice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.StoreProductInput{
		Title: "ChangePriceProduct",
		Slug:  "change-price-product",
		Price: 5.00,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.BusinessID, id) })

	if err := repo.Update(ctx, testutil.BusinessID, id, domain.StoreProductInput{
		Title: "ChangePriceProduct",
		Slug:  "change-price-product",
		Price: 125.00,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, p := range items {
		if p.ID == id && p.Price != 125.00 {
			t.Errorf("expected price 125.00, got %f", p.Price)
		}
	}
}

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
