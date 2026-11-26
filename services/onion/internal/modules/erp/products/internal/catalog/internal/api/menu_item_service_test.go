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
	if item.OwnerId != 0 || item.CategoryId != 0 || item.Sku != "" || item.Barcode != "" || item.TaxRate != 0 || item.Stock != 0 {
		t.Fatalf("public menu exposed internal or inventory fields: %+v", item)
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
