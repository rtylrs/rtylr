package products

import (
	"testing"

	"github.com/voxire/onion/internal/modules/products/internal/domain"
)

func TestParseGetProductToolInputAcceptsID(t *testing.T) {
	parsed, err := parseGetProductToolInput(map[string]any{"id": "42"})
	if err != nil {
		t.Fatalf("parseGetProductToolInput: %v", err)
	}
	if parsed.ProductID != 42 {
		t.Fatalf("ProductID = %d", parsed.ProductID)
	}
	if parsed.Barcode != "" {
		t.Fatalf("Barcode = %q", parsed.Barcode)
	}
}

func TestParseGetProductToolInputAcceptsBarcode(t *testing.T) {
	parsed, err := parseGetProductToolInput(map[string]any{"barcode": " 0123456789 "})
	if err != nil {
		t.Fatalf("parseGetProductToolInput: %v", err)
	}
	if parsed.ProductID != 0 {
		t.Fatalf("ProductID = %d", parsed.ProductID)
	}
	if parsed.Barcode != "0123456789" {
		t.Fatalf("Barcode = %q", parsed.Barcode)
	}
}

func TestParseGetProductToolInputRequiresIdentifier(t *testing.T) {
	_, err := parseGetProductToolInput(map[string]any{"name": "Protein Bar"})
	if err == nil {
		t.Fatal("expected missing identifier error")
	}
}

func TestProductToolPayloadIncludesBusinessFields(t *testing.T) {
	payload := productToolPayload(domain.Product{
		ID:                7,
		Name:              "Protein Bar",
		SKU:               "BAR-001",
		Barcode:           "0123456789",
		CostPrice:         1.25,
		SalePrice:         3.50,
		StockQuantity:     4,
		LowStockThreshold: 5,
		ReorderQuantity:   24,
		IsActive:          true,
		TrackInventory:    true,
	}, []domain.ProductAttribute{
		{ID: 3, Name: "Flavor", Value: "Chocolate", Position: 1},
	})

	if payload["gross_profit"] != 2.25 {
		t.Fatalf("gross_profit = %v", payload["gross_profit"])
	}
	if payload["stock_value"] != 5.0 {
		t.Fatalf("stock_value = %v", payload["stock_value"])
	}
	if payload["is_low_stock"] != true {
		t.Fatalf("is_low_stock = %v", payload["is_low_stock"])
	}
	if payload["status"] != "active" {
		t.Fatalf("status = %v", payload["status"])
	}
	attrs, ok := payload["attributes"].([]map[string]any)
	if !ok {
		t.Fatalf("attributes type = %T", payload["attributes"])
	}
	if len(attrs) != 1 || attrs[0]["name"] != "Flavor" {
		t.Fatalf("attributes = %#v", attrs)
	}
}

func TestParseCreateProductToolInput(t *testing.T) {
	parsed, err := parseCreateProductToolInput(map[string]any{
		"product_name":      " Protein Bar ",
		"sku":               " BAR-001 ",
		"barcode":           " 123456 ",
		"description":       "Chocolate protein bar",
		"price":             "3.50",
		"cost":              "1.25",
		"stock":             12,
		"low_stock":         "5",
		"reorder_qty":       "24",
		"category_id":       "10",
		"supplier_id":       11,
		"brand_id":          12,
		"currency":          "usd",
		"tax_rate":          "11",
		"is_active":         "yes",
		"track_inventory":   "true",
		"allow_backorder":   "no",
		"lead_time_days":    "7",
		"weight":            "0.08",
		"weight_unit":       "kg",
		"length":            "12",
		"width":             "4",
		"height":            "2",
		"dimension_unit":    "cm",
		"country_of_origin": "LB",
		"hs_code":           "210690",
		"compare_at_price":  "4.25",
	})
	if err != nil {
		t.Fatalf("parseCreateProductToolInput: %v", err)
	}
	if parsed.Name != "Protein Bar" {
		t.Fatalf("Name = %q", parsed.Name)
	}
	if parsed.SKU != "BAR-001" {
		t.Fatalf("SKU = %q", parsed.SKU)
	}
	if parsed.SalePrice != 3.50 {
		t.Fatalf("SalePrice = %v", parsed.SalePrice)
	}
	if parsed.CostPrice != 1.25 {
		t.Fatalf("CostPrice = %v", parsed.CostPrice)
	}
	if parsed.StockQuantity != 12 || !parsed.StockQuantitySet {
		t.Fatalf("StockQuantity = %v, StockQuantitySet = %v", parsed.StockQuantity, parsed.StockQuantitySet)
	}
	if parsed.LowStockThreshold != 5 {
		t.Fatalf("LowStockThreshold = %v", parsed.LowStockThreshold)
	}
	if parsed.ReorderQuantity != 24 {
		t.Fatalf("ReorderQuantity = %v", parsed.ReorderQuantity)
	}
	if parsed.CategoryID != 10 || parsed.SupplierID != 11 || parsed.BrandID != 12 {
		t.Fatalf("ids = category %d supplier %d brand %d", parsed.CategoryID, parsed.SupplierID, parsed.BrandID)
	}
	if parsed.Currency != "USD" {
		t.Fatalf("Currency = %q", parsed.Currency)
	}
	if !parsed.IsActive || !parsed.TrackInventory || parsed.AllowBackorder {
		t.Fatalf("booleans = active %v track %v backorder %v", parsed.IsActive, parsed.TrackInventory, parsed.AllowBackorder)
	}
	if parsed.LeadTimeDays != 7 {
		t.Fatalf("LeadTimeDays = %d", parsed.LeadTimeDays)
	}
}

func TestParseCreateProductToolInputDefaultsBooleans(t *testing.T) {
	parsed, err := parseCreateProductToolInput(map[string]any{"name": "Service Fee"})
	if err != nil {
		t.Fatalf("parseCreateProductToolInput: %v", err)
	}
	if !parsed.IsActive {
		t.Fatal("expected IsActive default true")
	}
	if !parsed.TrackInventory {
		t.Fatal("expected TrackInventory default true")
	}
	if parsed.AllowBackorder {
		t.Fatal("expected AllowBackorder default false")
	}
	if !parsed.StockQuantitySet {
		t.Fatal("expected StockQuantitySet true")
	}
}

func TestParseCreateProductToolInputRejectsMissingName(t *testing.T) {
	_, err := parseCreateProductToolInput(map[string]any{"price": 2})
	if err == nil {
		t.Fatal("expected missing name error")
	}
}

func TestParseCreateProductToolInputRejectsNegativePrice(t *testing.T) {
	_, err := parseCreateProductToolInput(map[string]any{"name": "Bad", "price": -1})
	if err == nil {
		t.Fatal("expected negative price error")
	}
}

func TestParseCreateProductToolInputRejectsBadBool(t *testing.T) {
	_, err := parseCreateProductToolInput(map[string]any{"name": "Bad", "is_active": "sometimes"})
	if err == nil {
		t.Fatal("expected bad bool error")
	}
}

func TestParseCreateProductToolInputRejectsBadCurrency(t *testing.T) {
	_, err := parseCreateProductToolInput(map[string]any{"name": "Bad", "currency": "US"})
	if err == nil {
		t.Fatal("expected bad currency error")
	}
}

func TestParseCreateProductToolInputPreservesImageURL(t *testing.T) {
	parsed, err := parseCreateProductToolInput(map[string]any{
		"name":       "Protein Bar",
		"sale_price": 3.50,
		"image_url":  "https://cdn.example.com/protein-bar.png",
	})
	if err != nil {
		t.Fatalf("parseCreateProductToolInput: %v", err)
	}
	if parsed.ImageURL != "https://cdn.example.com/protein-bar.png" {
		t.Fatalf("image URL = %q", parsed.ImageURL)
	}
}

func TestMergeProductUpdateToolInputPreservesUnchangedFields(t *testing.T) {
	parsed, err := mergeProductUpdateToolInput(domain.Product{
		ID:                42,
		Name:              "Protein Bar",
		SKU:               "BAR-001",
		CostPrice:         1.25,
		SalePrice:         3.50,
		StockQuantity:     12,
		LowStockThreshold: 5,
		ReorderQuantity:   24,
		IsActive:          true,
		TrackInventory:    true,
		AllowBackorder:    false,
		Currency:          "USD",
	}, map[string]any{
		"product_id": "42",
		"sale_price": 3.75,
	})
	if err != nil {
		t.Fatalf("mergeProductUpdateToolInput: %v", err)
	}
	if parsed.Name != "Protein Bar" || parsed.SKU != "BAR-001" {
		t.Fatalf("unchanged fields were not preserved: %+v", parsed)
	}
	if parsed.SalePrice != 3.75 || parsed.CostPrice != 1.25 {
		t.Fatalf("price fields = sale %v cost %v", parsed.SalePrice, parsed.CostPrice)
	}
	if parsed.StockQuantity != 12 || parsed.StockQuantitySet {
		t.Fatalf("stock fields = quantity %v set %v", parsed.StockQuantity, parsed.StockQuantitySet)
	}
}

func TestMergeProductUpdateToolInputAllowsExplicitStockChange(t *testing.T) {
	parsed, err := mergeProductUpdateToolInput(domain.Product{
		Name:          "Protein Bar",
		SalePrice:     3.50,
		CostPrice:     1.25,
		StockQuantity: 12,
	}, map[string]any{"product_id": 42, "quantity": 20})
	if err != nil {
		t.Fatalf("mergeProductUpdateToolInput: %v", err)
	}
	if parsed.StockQuantity != 20 || !parsed.StockQuantitySet {
		t.Fatalf("stock fields = quantity %v set %v", parsed.StockQuantity, parsed.StockQuantitySet)
	}
}

func TestMergeProductUpdateToolInputRejectsUnknownField(t *testing.T) {
	_, err := mergeProductUpdateToolInput(domain.Product{Name: "Protein Bar", SalePrice: 3.50}, map[string]any{
		"product_id": 42,
		"drop_table": true,
	})
	if err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestReorderSuggestionToolPayload(t *testing.T) {
	payload := reorderSuggestionToolPayload(domain.ReorderSuggestion{
		ProductID:    7,
		Name:         "Protein Bar",
		CurrentStock: 2,
		ReorderPoint: 5,
		SuggestedQty: 24,
		Confidence:   "low",
	})
	if payload["product_id"] != int64(7) || payload["suggested_quantity"] != float64(24) {
		t.Fatalf("payload = %#v", payload)
	}
}
