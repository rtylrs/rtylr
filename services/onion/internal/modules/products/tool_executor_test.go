package products

import "testing"

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
