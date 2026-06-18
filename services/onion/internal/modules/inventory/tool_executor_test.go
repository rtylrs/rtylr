package inventory

import "testing"

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
