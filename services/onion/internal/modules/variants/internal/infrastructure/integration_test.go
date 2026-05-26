package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/variants/internal/domain"
	"github.com/voxire/onion/internal/modules/variants/internal/infrastructure"
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
