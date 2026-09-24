package domain

import "testing"

func TestMergeProductUpdatePreservesOmittedFields(t *testing.T) {
	current := Product{
		CategoryID:        7,
		SupplierID:        8,
		Name:              "Existing name",
		SKU:               "SKU-1",
		Barcode:           "123",
		Description:       "Existing description",
		Unit:              "kg",
		CostPrice:         4.25,
		SalePrice:         9.50,
		StockQuantity:     23,
		LowStockThreshold: 5,
		ReorderQuantity:   10,
		IsActive:          true,
		TrackInventory:    true,
		BrandID:           11,
		CompareAtPrice:    12,
		TaxRate:           7.5,
		Currency:          "EUR",
		Weight:            1.25,
		WeightUnit:        "lb",
		Length:            3,
		Width:             2,
		Height:            4,
		DimensionUnit:     "in",
		CountryOfOrigin:   "DE",
		HSCode:            "1234",
		AllowBackorder:    true,
		LeadTimeDays:      6,
		ImageURL:          "https://cdn.example/existing.png",
	}
	input := ProductInput{
		Name:             "Updated name",
		SalePrice:        10.25,
		PresentFields:    map[string]bool{ProductFieldName: true, ProductFieldSalePrice: true},
		StockQuantitySet: false,
		ImageURLSet:      false,
	}

	merged := MergeProductUpdate(current, input)

	if merged.Name != "Updated name" || merged.SalePrice != 10.25 {
		t.Fatalf("supplied fields were not applied: name=%q sale_price=%v", merged.Name, merged.SalePrice)
	}
	if merged.StockQuantity != current.StockQuantity || merged.StockQuantitySet {
		t.Fatalf("omitted stock was changed: quantity=%v set=%v", merged.StockQuantity, merged.StockQuantitySet)
	}
	if !merged.IsActive || !merged.TrackInventory || merged.TaxRate != current.TaxRate || merged.Currency != current.Currency {
		t.Fatalf("omitted optional fields were not preserved: active=%v track=%v tax=%v currency=%q", merged.IsActive, merged.TrackInventory, merged.TaxRate, merged.Currency)
	}
	if merged.ImageURL != current.ImageURL {
		t.Fatalf("omitted image was changed: %q", merged.ImageURL)
	}
}

func TestMergeProductUpdateKeepsLegacyNonEmptyImageReplacement(t *testing.T) {
	merged := MergeProductUpdate(Product{}, ProductInput{ImageURL: "https://cdn.example/new.png"})

	if !merged.ImageURLSet {
		t.Fatal("legacy non-empty image replacement was not marked as present")
	}
}

func TestMergeProductUpdateHonorsExplicitZeroAndFalse(t *testing.T) {
	current := Product{StockQuantity: 12, IsActive: true, ImageURL: "old"}
	input := ProductInput{
		PresentFields: map[string]bool{
			ProductFieldStockQuantity: true,
			ProductFieldIsActive:      true,
			ProductFieldImageURL:      true,
		},
		StockQuantitySet: true,
		ImageURLSet:      true,
	}

	merged := MergeProductUpdate(current, input)

	if merged.StockQuantity != 0 || merged.IsActive || merged.ImageURL != "" {
		t.Fatalf("explicit zero/false/clear values were not honored: stock=%v active=%v image=%q", merged.StockQuantity, merged.IsActive, merged.ImageURL)
	}
}
