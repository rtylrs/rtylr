package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/store_products/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/store_products/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

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

// Regression: a stale editor (the storefront row was deleted in another tab)
// must not be able to rewrite the shared ERP catalog row and still be told the
// update succeeded. Update has to report the missing storefront row and leave
// erp_products untouched.
func TestUpdateStoreProduct_MissingRowReportsNotFoundAndLeavesCatalogIntact(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const businessID int64 = 9200180

	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM erp_product_images WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM store_products WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM erp_products WHERE business_id = ?", businessID)
	})

	catalogID := insertCatalogProduct(ctx, t, db, businessID, "Catalog before update")

	result, err := db.ExecContext(ctx, `
		INSERT INTO store_products (owner_id, business_id, product_id, title, price, sort_order)
		VALUES (?, ?, ?, 'Store row to delete', 10, 1)
	`, testutil.OwnerID, businessID, catalogID)
	if err != nil {
		t.Fatalf("insert store product: %v", err)
	}
	storeID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("store product id: %v", err)
	}

	// The stale list the caller is holding still references storeID.
	if _, err := db.ExecContext(ctx, "DELETE FROM store_products WHERE business_id = ? AND id = ?", businessID, storeID); err != nil {
		t.Fatalf("delete store product: %v", err)
	}

	err = repo.Update(ctx, businessID, storeID, domain.StoreProductInput{
		ProductID: catalogID,
		Title:     "Catalog after stale update",
		Slug:      "stale-update",
		Price:     999.99,
	})
	if !errors.Is(err, domain.ErrStoreProductNotFound) {
		t.Fatalf("Update on deleted store product error = %v, want ErrStoreProductNotFound", err)
	}

	var name, salePrice string
	if err := db.QueryRowContext(ctx,
		"SELECT name, sale_price FROM erp_products WHERE business_id = ? AND id = ?",
		businessID, catalogID,
	).Scan(&name, &salePrice); err != nil {
		t.Fatalf("read catalog product: %v", err)
	}
	if name != "Catalog before update" || salePrice != "10.00" {
		t.Fatalf("stale update mutated the ERP catalog row: name=%q sale_price=%q", name, salePrice)
	}
}
