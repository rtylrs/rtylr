package api

import (
	"testing"

	"github.com/voxire/onion/internal/modules/erp/products/internal/catalog/internal/domain"
)

func TestToPublicProtoMenuItemOmitsInventoryAndInternalIdentifiers(t *testing.T) {
	item := toPublicProtoMenuItem(domain.MenuItem{
		ID:          42,
		OwnerID:     7,
		BusinessID:  9,
		CategoryID:  3,
		Name:        "Coffee",
		SKU:         "COFFEE-01",
		Barcode:     "123456",
		SalePrice:   3.5,
		TaxRate:     0.15,
		Description: "Freshly brewed",
		ImageURL:    "https://example.test/coffee.jpg",
		Stock:       99,
		IsActive:    true,
	})

	if item.Id != 42 || item.Name != "Coffee" || item.SalePrice != 3.5 {
		t.Fatalf("public menu lost customer-facing fields: %+v", item)
	}
	if item.OwnerId != 0 || item.CategoryId != 0 || item.Sku != "" || item.Barcode != "" || item.TaxRate != 0 || item.Stock != 0 || item.TrackInventory != nil || item.AllowBackorder != nil {
		t.Fatalf("public menu exposed internal or inventory fields: %+v", item)
	}
}

func TestToPublicProtoMenuItemReflectsTrackedStockAvailability(t *testing.T) {
	cases := []struct {
		name           string
		item           domain.MenuItem
		expectedActive bool
	}{
		{
			name:           "tracked item with zero stock is unavailable",
			item:           domain.MenuItem{ID: 1, Name: "Croissant", IsActive: true, TrackInventory: true, Stock: 0},
			expectedActive: false,
		},
		{
			name:           "tracked item with negative stock is unavailable",
			item:           domain.MenuItem{ID: 2, Name: "Matcha", IsActive: true, TrackInventory: true, Stock: -2},
			expectedActive: false,
		},
		{
			name:           "tracked item with positive stock stays available",
			item:           domain.MenuItem{ID: 3, Name: "Espresso", IsActive: true, TrackInventory: true, Stock: 5},
			expectedActive: true,
		},
		{
			name:           "tracked item with zero stock and backorder enabled stays available",
			item:           domain.MenuItem{ID: 4, Name: "Custom Cake", IsActive: true, TrackInventory: true, AllowBackorder: true, Stock: 0},
			expectedActive: true,
		},
		{
			name:           "untracked item with zero stock stays available",
			item:           domain.MenuItem{ID: 5, Name: "House Tea", IsActive: true, TrackInventory: false, Stock: 0},
			expectedActive: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := toPublicProtoMenuItem(tc.item)
			if got.IsActive != tc.expectedActive {
				t.Fatalf("expected IsActive=%v, got %v", tc.expectedActive, got.IsActive)
			}
			if got.Stock != 0 || got.TrackInventory != nil || got.AllowBackorder != nil {
				t.Fatalf("public menu must not leak raw inventory fields: %+v", got)
			}
		})
	}
}

func TestSanitizeMenuImageURLRejectsUnsafeSchemes(t *testing.T) {
	valid := []string{"", "https://cdn.voxire.com/a.jpg", "/storage/uploads/1/a.png"}
	for _, raw := range valid {
		if _, err := sanitizeMenuImageURL(raw); err != nil {
			t.Fatalf("expected %q to be allowed, got %v", raw, err)
		}
	}
	invalid := []string{
		"javascript:alert(1)",
		"data:text/html,<script>alert(1)</script>",
		"//evil.example/a.png",
		"https://cdn.example.com/a\n.png",
	}
	for _, raw := range invalid {
		if _, err := sanitizeMenuImageURL(raw); err == nil {
			t.Fatalf("expected %q to be rejected", raw)
		}
	}
}
