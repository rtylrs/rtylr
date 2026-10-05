package inventory

import (
	"encoding/json"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/inventory/internal/domain"
)

func TestParseListProductsToolInput(t *testing.T) {
	got, err := parseListProductsToolInput(map[string]any{
		"query":          " whey ",
		"low_stock_only": "yes",
		"page":           json.Number("2"),
		"limit":          "10",
	})
	if err != nil {
		t.Fatalf("parseListProductsToolInput: %v", err)
	}
	if got.Search != "whey" {
		t.Fatalf("Search = %q", got.Search)
	}
	if !got.LowStockOnly {
		t.Fatal("expected low stock only")
	}
	if got.Page != 2 {
		t.Fatalf("Page = %d", got.Page)
	}
	if got.PageSize != 10 {
		t.Fatalf("PageSize = %d", got.PageSize)
	}
}

func TestParseListProductsToolInputDefaultsAndCapsPageSize(t *testing.T) {
	got, err := parseListProductsToolInput(map[string]any{
		"page_size": 500,
	})
	if err != nil {
		t.Fatalf("parseListProductsToolInput: %v", err)
	}
	if got.Page != 1 {
		t.Fatalf("Page = %d", got.Page)
	}
	if got.PageSize != maxListProductsToolPageSize {
		t.Fatalf("PageSize = %d", got.PageSize)
	}
}

func TestParseListProductsToolInputRejectsInvalidBool(t *testing.T) {
	_, err := parseListProductsToolInput(map[string]any{
		"low_stock_only": "maybe",
	})
	if err == nil {
		t.Fatal("expected invalid bool error")
	}
}

func TestFilterInventoryReportItems(t *testing.T) {
	items := []domain.InventoryReportItem{
		{ProductID: 1, ProductName: "Whey Protein", SKU: "WP-1", IsLowStock: true},
		{ProductID: 2, ProductName: "Creatine", SKU: "CR-1", IsLowStock: true},
		{ProductID: 3, ProductName: "Mass Gainer", SKU: "MG-1", IsLowStock: false},
	}

	got := filterInventoryReportItems(items, listProductsToolInput{
		Search:       "wp",
		LowStockOnly: true,
	})
	if len(got) != 1 {
		t.Fatalf("len = %d", len(got))
	}
	if got[0].ProductID != 1 {
		t.Fatalf("ProductID = %d", got[0].ProductID)
	}
}

func TestPaginateInventoryReportItems(t *testing.T) {
	items := []domain.InventoryReportItem{
		{ProductID: 1},
		{ProductID: 2},
		{ProductID: 3},
	}
	got := paginateInventoryReportItems(items, 2, 2)
	if len(got) != 1 {
		t.Fatalf("len = %d", len(got))
	}
	if got[0].ProductID != 3 {
		t.Fatalf("ProductID = %d", got[0].ProductID)
	}
}

func TestParseSetStockToolInputAcceptsQuantityAliases(t *testing.T) {
	got, err := parseSetStockToolInput(map[string]any{
		"product_id":     "5201",
		"stock_quantity": "12.5",
		"notes":          " cycle count ",
	})
	if err != nil {
		t.Fatalf("parseSetStockToolInput: %v", err)
	}
	if got.ProductID != 5201 {
		t.Fatalf("product id = %d", got.ProductID)
	}
	if got.Quantity != 12.5 {
		t.Fatalf("quantity = %v", got.Quantity)
	}
	if got.Notes != "cycle count" {
		t.Fatalf("notes = %q", got.Notes)
	}
}

func TestParseSetStockToolInputRejectsInvalidValues(t *testing.T) {
	cases := []map[string]any{
		{"quantity": 4},
		{"product_id": 1},
		{"product_id": 1, "quantity": nil},
		{"product_id": 1, "quantity": -1},
		{"product_id": 1.5, "quantity": 1},
		{"product_id": 1, "quantity": "oops"},
	}
	for i, input := range cases {
		if _, err := parseSetStockToolInput(input); err == nil {
			t.Fatalf("case %d: expected validation error", i)
		}
	}
}
