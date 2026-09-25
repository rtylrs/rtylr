package infrastructure_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"slices"
	"testing"

	"github.com/voxire/onion/internal/modules/store_products/internal/domain"
	"github.com/voxire/onion/internal/modules/store_products/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestStoreProducts_List_CatalogImagesAreScopedToStorePage(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const businessID int64 = 9200172
	const otherBusinessID int64 = 9200173

	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM erp_product_images WHERE business_id IN (?, ?)", businessID, otherBusinessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM store_products WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM erp_products WHERE business_id = ?", businessID)
	})

	firstProductID := insertCatalogProduct(ctx, t, db, businessID, "First catalog product")
	secondProductID := insertCatalogProduct(ctx, t, db, businessID, "Second catalog product")
	insertStoreProduct(ctx, t, db, businessID, firstProductID, "First store product", 1)
	insertStoreProduct(ctx, t, db, businessID, secondProductID, "Second store product", 2)

	for _, image := range []struct {
		businessID int64
		productID  int64
		url        string
		position   int
	}{
		{businessID, firstProductID, "https://images.example/first-1.jpg", 1},
		{businessID, firstProductID, "https://images.example/first-2.jpg", 2},
		{businessID, secondProductID, "https://images.example/second.jpg", 1},
		{otherBusinessID, firstProductID, "https://images.example/cross-tenant.jpg", 1},
	} {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO erp_product_images (owner_id, business_id, product_id, url, position)
			VALUES (?, ?, ?, ?, ?)
		`, testutil.OwnerID, image.businessID, image.productID, image.url, image.position); err != nil {
			t.Fatalf("insert catalog image: %v", err)
		}
	}

	items, err := repo.List(ctx, businessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected two store products, got %d", len(items))
	}
	if items[0].ProductID != firstProductID || items[1].ProductID != secondProductID {
		t.Fatalf("expected stable store ordering, got product ids %d then %d", items[0].ProductID, items[1].ProductID)
	}
	assertStoreProductImages(t, items[0].Images, []string{
		"https://images.example/first-1.jpg",
		"https://images.example/first-2.jpg",
	})
	assertStoreProductImages(t, items[1].Images, []string{"https://images.example/second.jpg"})
}

func assertStoreProductImages(t *testing.T, got string, want []string) {
	t.Helper()
	var actual []string
	if err := json.Unmarshal([]byte(got), &actual); err != nil {
		t.Fatalf("decode store product images %q: %v", got, err)
	}
	slices.Sort(actual)
	slices.Sort(want)
	if !slices.Equal(actual, want) {
		t.Fatalf("expected images %v, got %v", want, actual)
	}
}

func insertCatalogProduct(ctx context.Context, t *testing.T, db *sql.DB, businessID int64, name string) int64 {
	t.Helper()
	result, err := db.ExecContext(ctx, `
		INSERT INTO erp_products (owner_id, business_id, name, sale_price)
		VALUES (?, ?, ?, 10)
	`, testutil.OwnerID, businessID, name)
	if err != nil {
		t.Fatalf("insert catalog product: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("catalog product id: %v", err)
	}
	return id
}

func insertStoreProduct(ctx context.Context, t *testing.T, db *sql.DB, businessID, productID int64, title string, sortOrder int) {
	t.Helper()
	_, err := db.ExecContext(ctx, `
		INSERT INTO store_products (owner_id, business_id, product_id, title, price, sort_order)
		VALUES (?, ?, ?, ?, 10, ?)
	`, testutil.OwnerID, businessID, productID, title, sortOrder)
	if err != nil {
		t.Fatalf("insert store product: %v", err)
	}
}

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
