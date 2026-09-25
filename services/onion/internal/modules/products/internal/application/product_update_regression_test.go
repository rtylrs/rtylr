package application_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/products/internal/application"
	"github.com/voxire/onion/internal/modules/products/internal/domain"
	"github.com/voxire/onion/internal/modules/products/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestUpdateProduct_PresenceAwareInputPreservesOmittedFields(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	service := application.NewService(context.Background(), db)
	ctx := context.Background()

	id, err := repo.CreateProduct(ctx, testutil.OwnerID, domain.ProductInput{
		Name:              "Before update",
		SKU:               fmt.Sprintf("PRESENCE-%d", time.Now().UnixNano()),
		CostPrice:         4,
		SalePrice:         8,
		LowStockThreshold: 4,
		ReorderQuantity:   8,
		IsActive:          true,
		TrackInventory:    true,
		CompareAtPrice:    10,
		TaxRate:           9.5,
		Currency:          "EUR",
		Weight:            1.5,
		WeightUnit:        "lb",
		Length:            2,
		Width:             3,
		Height:            4,
		DimensionUnit:     "in",
		CountryOfOrigin:   "DE",
		HSCode:            "1234",
		AllowBackorder:    true,
		LeadTimeDays:      5,
	})
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM erp_product_images WHERE business_id = ? AND product_id = ?`, testutil.OwnerID, id)
		_ = repo.DeleteProduct(ctx, testutil.OwnerID, id)
	})

	err = service.UpdateProduct(ctx, testutil.OwnerID, id, domain.ProductInput{
		Name:      "After update",
		SalePrice: 11,
		PresentFields: map[string]bool{
			domain.ProductFieldName:      true,
			domain.ProductFieldSalePrice: true,
		},
	})
	if err != nil {
		t.Fatalf("UpdateProduct: %v", err)
	}

	got, err := repo.GetProduct(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetProduct: %v", err)
	}
	if got.Name != "After update" || got.SalePrice != 11 {
		t.Fatalf("supplied fields not applied: name=%q sale_price=%v", got.Name, got.SalePrice)
	}
	if !got.IsActive || !got.TrackInventory || got.TaxRate != 9.5 || got.Currency != "EUR" || got.Weight != 1.5 || got.WeightUnit != "lb" {
		t.Fatalf("omitted optional fields changed: active=%v track=%v tax=%v currency=%q weight=%v weight_unit=%q", got.IsActive, got.TrackInventory, got.TaxRate, got.Currency, got.Weight, got.WeightUnit)
	}
	if got.Length != 2 || got.Width != 3 || got.Height != 4 || got.DimensionUnit != "in" || got.CountryOfOrigin != "DE" || got.HSCode != "1234" || !got.AllowBackorder || got.LeadTimeDays != 5 {
		t.Fatalf("omitted PIM fields changed: length=%v width=%v height=%v dimension=%q country=%q hs=%q backorder=%v lead=%d", got.Length, got.Width, got.Height, got.DimensionUnit, got.CountryOfOrigin, got.HSCode, got.AllowBackorder, got.LeadTimeDays)
	}
}
