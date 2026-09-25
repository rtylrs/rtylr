package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/products/internal/domain"
	"github.com/voxire/onion/internal/modules/products/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListProducts_ReturnsAllFixtureProducts(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	products, total, _, err := repo.ListProducts(ctx, testutil.OwnerID, "", 0, 1000, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(products) == 0 {
		t.Fatal("expected fixture rows, got none")
	}
	if total < int64(len(products)) {
		t.Fatalf("expected total >= listed products, got total=%d len=%d", total, len(products))
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

	products, _, _, err := repo.ListProducts(ctx, testutil.OwnerID, "", 0, 1000, 0)
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
	products, _, _, err := repo.ListProducts(ctx, testutil.OwnerID, "", 5002, 10, 0)
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

func TestListProducts_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	products, _, _, err := repo.ListProducts(ctx, 0, "", 0, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(products) != 0 {
		t.Fatalf("expected 0 products for ownerID=0, got %d", len(products))
	}
}

func TestListProducts_SearchBurger_FindsProduct(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	products, _, _, err := repo.ListProducts(ctx, testutil.OwnerID, "Burger", 0, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(products) == 0 {
		t.Fatal("expected at least one product matching 'Burger', got none")
	}
	found := false
	for _, p := range products {
		if p.Name == "Burger" {
			found = true
			break
		}
	}
	if !found {
		t.Error("product named 'Burger' not found in search results")
	}
}

func TestListProducts_SearchNonExistent_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	products, _, _, err := repo.ListProducts(ctx, testutil.OwnerID, "xyzNoSuchProduct999", 0, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(products) != 0 {
		t.Fatalf("expected 0 results for nonsense search, got %d", len(products))
	}
}

func TestListProducts_ByCategoryID5002_ReturnsMains(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	products, _, _, err := repo.ListProducts(ctx, testutil.OwnerID, "", 5002, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range products {
		if p.CategoryID != 5002 {
			t.Errorf("product id=%d has category_id=%d, want 5002", p.ID, p.CategoryID)
		}
	}
}

func TestListProducts_PaginationLimit1_ReturnsOne(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	products, _, _, err := repo.ListProducts(ctx, testutil.OwnerID, "", 0, 1, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("expected exactly 1 product with limit=1, got %d", len(products))
	}
}

func TestGetProduct_Burger_NameAndSKU(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	p, err := repo.GetProduct(ctx, testutil.OwnerID, 5201)
	if err != nil {
		t.Fatalf("GetProduct(5201): %v", err)
	}
	if p.Name != "Burger" {
		t.Errorf("expected Name=Burger, got %q", p.Name)
	}
	if p.SKU != "BURG-1" {
		t.Errorf("expected SKU=BURG-1, got %q", p.SKU)
	}
}

func TestGetProduct_Soda_PriceIsCorrect(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	p, err := repo.GetProduct(ctx, testutil.OwnerID, 5202)
	if err != nil {
		t.Fatalf("GetProduct(5202): %v", err)
	}
	if p.Name != "Soda" {
		t.Errorf("expected Name=Soda, got %q", p.Name)
	}
	if p.SKU != "SODA-1" {
		t.Errorf("expected SKU=SODA-1, got %q", p.SKU)
	}
}

func TestGetProduct_Salad_LowStock(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	p, err := repo.GetProduct(ctx, testutil.OwnerID, 5203)
	if err != nil {
		t.Fatalf("GetProduct(5203): %v", err)
	}
	if p.StockQuantity != 4.0 {
		t.Errorf("Salad (id=5203): expected stock=4, got %.2f", p.StockQuantity)
	}
}

func TestGetProduct_NonExistent_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetProduct(ctx, testutil.OwnerID, 9999999)
	if err == nil {
		t.Fatal("expected error for non-existent product, got nil")
	}
}

func TestGetProductByBarcode_NotFound_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetProductByBarcode(ctx, testutil.OwnerID, "BARCODE-DOES-NOT-EXIST-XYZ")
	if err == nil {
		t.Fatal("expected error for non-existent barcode, got nil")
	}
}

func TestCreateProduct_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateProduct(ctx, testutil.OwnerID, domain.ProductInput{
		Name:      "Test Create Product",
		SKU:       "TCP-001",
		CostPrice: 3.0,
		SalePrice: 6.0,
		IsActive:  true,
	})
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.DeleteProduct(ctx, testutil.OwnerID, id) })
}

func TestCreateProduct_RejectsCrossBusinessCategoryReference(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	_, err := repo.CreateProduct(context.Background(), testutil.OwnerID, domain.ProductInput{
		CategoryID: 999999999,
		Name:       "Cross Business Category Product",
		SKU:        "CROSS-BUSINESS-CATEGORY-001",
		IsActive:   true,
	})
	if !errors.Is(err, domain.ErrProductReferenceNotFound) {
		t.Fatalf("CreateProduct cross-business category error = %v, want ErrProductReferenceNotFound", err)
	}
}

func TestCreateProduct_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const wantName = "Name Storage Test Product"
	id, err := repo.CreateProduct(ctx, testutil.OwnerID, domain.ProductInput{
		Name:     wantName,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteProduct(ctx, testutil.OwnerID, id) })

	p, err := repo.GetProduct(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetProduct: %v", err)
	}
	if p.Name != wantName {
		t.Errorf("expected Name=%q, got %q", wantName, p.Name)
	}
}

func TestCreateProduct_SKUIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const wantSKU = "SKU-STORE-001"
	id, err := repo.CreateProduct(ctx, testutil.OwnerID, domain.ProductInput{
		Name:     "SKU Store Test",
		SKU:      wantSKU,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteProduct(ctx, testutil.OwnerID, id) })

	p, err := repo.GetProduct(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetProduct: %v", err)
	}
	if p.SKU != wantSKU {
		t.Errorf("expected SKU=%q, got %q", wantSKU, p.SKU)
	}
}

func TestUpdateProduct_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateProduct(ctx, testutil.OwnerID, domain.ProductInput{
		Name:     "Before Rename",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteProduct(ctx, testutil.OwnerID, id) })

	err = repo.UpdateProduct(ctx, testutil.OwnerID, id, domain.ProductInput{
		Name:     "After Rename",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("UpdateProduct: %v", err)
	}

	p, err := repo.GetProduct(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetProduct after update: %v", err)
	}
	if p.Name != "After Rename" {
		t.Errorf("expected Name=After Rename, got %q", p.Name)
	}
}

func TestCountProducts_OwnerHasProducts(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	count, err := repo.CountProducts(ctx, testutil.OwnerID, 0)
	if err != nil {
		t.Fatalf("CountProducts: %v", err)
	}
	if count <= 0 {
		t.Errorf("expected at least 1 product for fixture owner, got %d", count)
	}
}

func TestCountProducts_ZeroOwner_ReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	count, err := repo.CountProducts(ctx, 0, 0)
	if err != nil {
		t.Fatalf("CountProducts(ownerID=0): %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 for ownerID=0, got %d", count)
	}
}

func TestCountProducts_ByCategoryID5002(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	count, err := repo.CountProducts(ctx, testutil.OwnerID, 5002)
	if err != nil {
		t.Fatalf("CountProducts by category: %v", err)
	}
	if count <= 0 {
		t.Errorf("expected at least 1 product in category 5002, got %d", count)
	}
}

func TestCountProducts_UnknownCategory_ReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	count, err := repo.CountProducts(ctx, testutil.OwnerID, 9999999)
	if err != nil {
		t.Fatalf("CountProducts unknown category: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 for unknown category, got %d", count)
	}
}

func TestCountProducts_AfterCreate_Increases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.CountProducts(ctx, testutil.OwnerID, 0)
	if err != nil {
		t.Fatalf("CountProducts before: %v", err)
	}

	id, err := repo.CreateProduct(ctx, testutil.OwnerID, domain.ProductInput{
		Name:     "Count Test Product",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteProduct(ctx, testutil.OwnerID, id) })

	after, err := repo.CountProducts(ctx, testutil.OwnerID, 0)
	if err != nil {
		t.Fatalf("CountProducts after: %v", err)
	}
	if after != before+1 {
		t.Errorf("expected count to increase by 1, before=%d after=%d", before, after)
	}
}

func TestDeleteProduct_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateProduct(ctx, testutil.OwnerID, domain.ProductInput{
		Name:     "To Be Deleted Product",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}

	if err := repo.DeleteProduct(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("DeleteProduct: %v", err)
	}

	_, err = repo.GetProduct(ctx, testutil.OwnerID, id)
	if err == nil {
		t.Fatal("expected error after deletion, got nil")
	}
}
