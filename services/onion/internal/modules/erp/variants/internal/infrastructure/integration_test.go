package infrastructure_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/variants/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/variants/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const variantsProductID int64 = 5201

func TestVariants_ListVariants_AllOwner_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.ListVariants(context.Background(), testutil.OwnerID, 0)
	if err != nil {
		t.Fatalf("ListVariants (all): %v", err)
	}
}

func TestVariants_ListVariants_ByProduct_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.ListVariants(context.Background(), testutil.OwnerID, variantsProductID)
	if err != nil {
		t.Fatalf("ListVariants by product: %v", err)
	}
}

func TestVariants_ListVariants_OwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	variants, err := repo.ListVariants(context.Background(), testutil.OwnerID, 0)
	if err != nil {
		t.Fatalf("ListVariants: %v", err)
	}
	for _, v := range variants {
		if v.OwnerID != testutil.OwnerID {
			t.Errorf("expected owner_id %d, got %d", testutil.OwnerID, v.OwnerID)
		}
	}
}

func TestVariants_CreateVariant_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateVariant(context.Background(), testutil.OwnerID, domain.VariantInput{
		ProductID:     variantsProductID,
		Name:          "Small Blue",
		SKU:           "SB-001",
		CostPrice:     5.0,
		SalePrice:     10.0,
		StockQuantity: 100.0,
		IsActive:      true,
	})
	if err != nil {
		t.Fatalf("CreateVariant: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.DeleteVariant(context.Background(), testutil.OwnerID, id) })
}

func TestCreateVariant_RejectsForeignBusinessProduct(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	foreignBusinessID := testutil.OwnerID + 9000

	result, err := db.ExecContext(ctx, `
		INSERT INTO erp_products (owner_id, business_id, name)
		VALUES (?, ?, 'Foreign product')`, testutil.OwnerID+9001, foreignBusinessID)
	if err != nil {
		t.Fatalf("insert foreign product: %v", err)
	}
	foreignProductID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("foreign product id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM erp_products WHERE id = ?", foreignProductID)
	})

	id, err := repo.CreateVariant(ctx, testutil.OwnerID, domain.VariantInput{
		ProductID: foreignProductID, Name: "Cross-tenant variant", IsActive: true,
	})
	if !errors.Is(err, domain.ErrVariantProductNotFound) {
		t.Fatalf("CreateVariant foreign product id=%d err=%v, want ErrVariantProductNotFound", id, err)
	}
}

func TestVariants_CreateVariant_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateVariant(context.Background(), testutil.OwnerID, domain.VariantInput{
		ProductID: variantsProductID,
		Name:      "List Test Variant",
		SalePrice: 15.0,
		IsActive:  true,
	})
	if err != nil {
		t.Fatalf("CreateVariant: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteVariant(context.Background(), testutil.OwnerID, id) })

	variants, err := repo.ListVariants(context.Background(), testutil.OwnerID, variantsProductID)
	if err != nil {
		t.Fatalf("ListVariants: %v", err)
	}
	found := false
	for _, v := range variants {
		if v.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("created variant id=%d not found in list", id)
	}
}

func TestVariants_UpdateVariant_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateVariant(context.Background(), testutil.OwnerID, domain.VariantInput{
		ProductID: variantsProductID, Name: "OldVariantName", SalePrice: 5.0, IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateVariant: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteVariant(context.Background(), testutil.OwnerID, id) })

	if err := repo.UpdateVariant(context.Background(), testutil.OwnerID, id, domain.VariantInput{
		ProductID: variantsProductID, Name: "NewVariantName", SalePrice: 5.0, IsActive: true,
	}); err != nil {
		t.Fatalf("UpdateVariant: %v", err)
	}

	variants, _ := repo.ListVariants(context.Background(), testutil.OwnerID, variantsProductID)
	for _, v := range variants {
		if v.ID == id && v.Name != "NewVariantName" {
			t.Errorf("expected name NewVariantName, got %s", v.Name)
		}
	}
}

func TestVariants_DeleteVariant_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateVariant(context.Background(), testutil.OwnerID, domain.VariantInput{
		ProductID: variantsProductID, Name: "DeleteMe", SalePrice: 1.0, IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateVariant: %v", err)
	}

	if err := repo.DeleteVariant(context.Background(), testutil.OwnerID, id); err != nil {
		t.Fatalf("DeleteVariant: %v", err)
	}

	variants, _ := repo.ListVariants(context.Background(), testutil.OwnerID, variantsProductID)
	for _, v := range variants {
		if v.ID == id {
			t.Errorf("deleted variant id=%d still in list", id)
		}
	}
}

func TestVariants_ListVariants_UnknownOwner_Empty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	variants, err := repo.ListVariants(context.Background(), 9999999, 0)
	if err != nil {
		t.Fatalf("ListVariants unknown owner: %v", err)
	}
	if len(variants) != 0 {
		t.Errorf("expected 0 for unknown owner, got %d", len(variants))
	}
}

func TestVariants_CreateVariant_WithBarcode(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateVariant(context.Background(), testutil.OwnerID, domain.VariantInput{
		ProductID: variantsProductID,
		Name:      "Barcoded Variant",
		Barcode:   "1234567890123",
		SalePrice: 20.0,
		IsActive:  true,
	})
	if err != nil {
		t.Fatalf("CreateVariant with barcode: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteVariant(context.Background(), testutil.OwnerID, id) })
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestVariants_UpdateVariant_Price(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateVariant(context.Background(), testutil.OwnerID, domain.VariantInput{
		ProductID: variantsProductID, Name: "PriceVariant", SalePrice: 10.0, IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateVariant: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteVariant(context.Background(), testutil.OwnerID, id) })

	if err := repo.UpdateVariant(context.Background(), testutil.OwnerID, id, domain.VariantInput{
		ProductID: variantsProductID, Name: "PriceVariant", SalePrice: 25.99, IsActive: true,
	}); err != nil {
		t.Fatalf("UpdateVariant price: %v", err)
	}
}

func TestListVariants_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	variants, err := repo.ListVariants(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("ListVariants(ownerID=0): %v", err)
	}
	if len(variants) != 0 {
		t.Fatalf("expected 0 variants for ownerID=0, got %d", len(variants))
	}
}

func TestListVariants_UnknownProductIsRejected(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.ListVariants(context.Background(), testutil.OwnerID, 9999999)
	if !errors.Is(err, domain.ErrVariantProductNotFound) {
		t.Fatalf("ListVariants unknown product error = %v, want ErrVariantProductNotFound", err)
	}
}

func TestCreateVariant_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateVariant(context.Background(), testutil.OwnerID, domain.VariantInput{
		ProductID: variantsProductID,
		Name:      "Persist Test Variant",
		SalePrice: 7.5,
		IsActive:  true,
	})
	if err != nil {
		t.Fatalf("CreateVariant: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.DeleteVariant(context.Background(), testutil.OwnerID, id) })
}

func TestCreateVariant_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	const wantName = "NameStoreVariant"
	id, err := repo.CreateVariant(context.Background(), testutil.OwnerID, domain.VariantInput{
		ProductID: variantsProductID,
		Name:      wantName,
		SalePrice: 9.0,
		IsActive:  true,
	})
	if err != nil {
		t.Fatalf("CreateVariant: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteVariant(context.Background(), testutil.OwnerID, id) })

	variants, err := repo.ListVariants(context.Background(), testutil.OwnerID, variantsProductID)
	if err != nil {
		t.Fatalf("ListVariants: %v", err)
	}
	for _, v := range variants {
		if v.ID == id {
			if v.Name != wantName {
				t.Errorf("expected Name=%q, got %q", wantName, v.Name)
			}
			return
		}
	}
	t.Fatalf("created variant id=%d not found in list", id)
}

func TestCreateVariant_ProductIDIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateVariant(context.Background(), testutil.OwnerID, domain.VariantInput{
		ProductID: variantsProductID,
		Name:      "ProductIDVariant",
		SalePrice: 5.0,
		IsActive:  true,
	})
	if err != nil {
		t.Fatalf("CreateVariant: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteVariant(context.Background(), testutil.OwnerID, id) })

	variants, err := repo.ListVariants(context.Background(), testutil.OwnerID, variantsProductID)
	if err != nil {
		t.Fatalf("ListVariants: %v", err)
	}
	for _, v := range variants {
		if v.ID == id {
			if v.ProductID != variantsProductID {
				t.Errorf("expected ProductID=%d, got %d", variantsProductID, v.ProductID)
			}
			return
		}
	}
	t.Fatalf("created variant id=%d not found in list", id)
}

func TestCreateVariant_PriceIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	const wantPrice = 42.99
	id, err := repo.CreateVariant(context.Background(), testutil.OwnerID, domain.VariantInput{
		ProductID: variantsProductID,
		Name:      "PriceStoreVariant",
		SalePrice: wantPrice,
		IsActive:  true,
	})
	if err != nil {
		t.Fatalf("CreateVariant: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteVariant(context.Background(), testutil.OwnerID, id) })

	variants, err := repo.ListVariants(context.Background(), testutil.OwnerID, variantsProductID)
	if err != nil {
		t.Fatalf("ListVariants: %v", err)
	}
	for _, v := range variants {
		if v.ID == id {
			if v.SalePrice != wantPrice {
				t.Errorf("expected SalePrice=%.2f, got %.2f", wantPrice, v.SalePrice)
			}
			return
		}
	}
	t.Fatalf("created variant id=%d not found in list", id)
}

func TestUpdateVariant_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateVariant(context.Background(), testutil.OwnerID, domain.VariantInput{
		ProductID: variantsProductID, Name: "OriginalVariantName", SalePrice: 5.0, IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateVariant: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteVariant(context.Background(), testutil.OwnerID, id) })

	if err := repo.UpdateVariant(context.Background(), testutil.OwnerID, id, domain.VariantInput{
		ProductID: variantsProductID, Name: "UpdatedVariantName", SalePrice: 5.0, IsActive: true,
	}); err != nil {
		t.Fatalf("UpdateVariant: %v", err)
	}

	variants, err := repo.ListVariants(context.Background(), testutil.OwnerID, variantsProductID)
	if err != nil {
		t.Fatalf("ListVariants: %v", err)
	}
	for _, v := range variants {
		if v.ID == id && v.Name != "UpdatedVariantName" {
			t.Errorf("expected Name=UpdatedVariantName, got %q", v.Name)
		}
	}
}

func TestUpdateVariant_ChangesPrice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateVariant(context.Background(), testutil.OwnerID, domain.VariantInput{
		ProductID: variantsProductID, Name: "ChangePriceVariant", SalePrice: 10.0, IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateVariant: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteVariant(context.Background(), testutil.OwnerID, id) })

	const newPrice = 99.99
	if err := repo.UpdateVariant(context.Background(), testutil.OwnerID, id, domain.VariantInput{
		ProductID: variantsProductID, Name: "ChangePriceVariant", SalePrice: newPrice, IsActive: true,
	}); err != nil {
		t.Fatalf("UpdateVariant: %v", err)
	}

	variants, err := repo.ListVariants(context.Background(), testutil.OwnerID, variantsProductID)
	if err != nil {
		t.Fatalf("ListVariants: %v", err)
	}
	for _, v := range variants {
		if v.ID == id {
			if v.SalePrice != newPrice {
				t.Errorf("expected SalePrice=%.2f, got %.2f", newPrice, v.SalePrice)
			}
			return
		}
	}
}

func TestDeleteVariant_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateVariant(context.Background(), testutil.OwnerID, domain.VariantInput{
		ProductID: variantsProductID, Name: "DeleteRowVariant", SalePrice: 1.0, IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateVariant: %v", err)
	}

	if err := repo.DeleteVariant(context.Background(), testutil.OwnerID, id); err != nil {
		t.Fatalf("DeleteVariant: %v", err)
	}

	variants, err := repo.ListVariants(context.Background(), testutil.OwnerID, variantsProductID)
	if err != nil {
		t.Fatalf("ListVariants after delete: %v", err)
	}
	for _, v := range variants {
		if v.ID == id {
			t.Errorf("deleted variant id=%d still appears in list", id)
		}
	}
}

func TestDeleteVariant_NonExistentReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	err := repo.DeleteVariant(context.Background(), testutil.OwnerID, 9999999)
	if !errors.Is(err, domain.ErrVariantNotFound) {
		t.Fatalf("DeleteVariant(non-existent) error = %v, want ErrVariantNotFound", err)
	}
}

func TestCreate_ThenList_FindsVariant(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateVariant(context.Background(), testutil.OwnerID, domain.VariantInput{
		ProductID: variantsProductID,
		Name:      "RoundtripVariant",
		SalePrice: 12.0,
		IsActive:  true,
	})
	if err != nil {
		t.Fatalf("CreateVariant: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteVariant(context.Background(), testutil.OwnerID, id) })

	variants, err := repo.ListVariants(context.Background(), testutil.OwnerID, variantsProductID)
	if err != nil {
		t.Fatalf("ListVariants: %v", err)
	}
	found := false
	for _, v := range variants {
		if v.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created variant id=%d not found in list", id)
	}
}

func TestListVariantsPage_ReturnsBoundedPagesAndTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	ids := make([]int64, 0, 3)
	for i := 0; i < 3; i++ {
		id, err := repo.CreateVariant(ctx, testutil.OwnerID, domain.VariantInput{
			ProductID: variantsProductID, Name: fmt.Sprintf("Pagination variant %d", i), IsActive: true,
		})
		if err != nil {
			t.Fatalf("CreateVariant[%d]: %v", i, err)
		}
		ids = append(ids, id)
	}
	t.Cleanup(func() {
		for _, id := range ids {
			_ = repo.DeleteVariant(context.Background(), testutil.OwnerID, id)
		}
	})

	first, hasMore, total, err := repo.ListVariantsPage(ctx, testutil.OwnerID, variantsProductID, 2, 0)
	if err != nil {
		t.Fatalf("first page: %v", err)
	}
	second, _, _, err := repo.ListVariantsPage(ctx, testutil.OwnerID, variantsProductID, 2, 2)
	if err != nil {
		t.Fatalf("second page: %v", err)
	}
	if len(first) != 2 || !hasMore || total < 3 {
		t.Fatalf("first page len=%d hasMore=%t total=%d", len(first), hasMore, total)
	}
	if len(second) == 0 || first[0].ID == second[0].ID {
		t.Fatalf("second page did not advance: first=%v second=%v", first, second)
	}
}

func TestCountVariants_AllOwner_NonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	count, err := repo.CountVariants(ctx, testutil.OwnerID, 0)
	if err != nil {
		t.Fatalf("CountVariants: %v", err)
	}
	if count < 0 {
		t.Errorf("expected non-negative count, got %d", count)
	}
}

func TestCountVariants_ZeroOwner_ReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	count, err := repo.CountVariants(ctx, 0, 0)
	if err != nil {
		t.Fatalf("CountVariants(ownerID=0): %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 for ownerID=0, got %d", count)
	}
}

func TestCountVariants_AfterCreate_Increases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.CountVariants(ctx, testutil.OwnerID, variantsProductID)
	if err != nil {
		t.Fatalf("CountVariants before: %v", err)
	}

	id, err := repo.CreateVariant(ctx, testutil.OwnerID, domain.VariantInput{
		ProductID: variantsProductID,
		Name:      "Count Test Variant",
		SalePrice: 5.0,
		IsActive:  true,
	})
	if err != nil {
		t.Fatalf("CreateVariant: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteVariant(ctx, testutil.OwnerID, id) })

	after, err := repo.CountVariants(ctx, testutil.OwnerID, variantsProductID)
	if err != nil {
		t.Fatalf("CountVariants after: %v", err)
	}
	if after != before+1 {
		t.Errorf("expected count to increase by 1, before=%d after=%d", before, after)
	}
}

func TestCountVariants_UnknownProduct_ReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	count, err := repo.CountVariants(ctx, testutil.OwnerID, 9999999)
	if err != nil {
		t.Fatalf("CountVariants unknown product: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 for unknown product, got %d", count)
	}
}

func TestListVariants_WrongOwnerRejectsForeignProduct(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.ListVariants(context.Background(), testutil.OwnerID+9999, variantsProductID)
	if !errors.Is(err, domain.ErrVariantProductNotFound) {
		t.Fatalf("ListVariants wrong owner error = %v, want ErrVariantProductNotFound", err)
	}
}
