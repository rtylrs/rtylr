package purchase_orders

import "testing"

func TestParseCreatePurchaseOrderToolInput(t *testing.T) {
	parsed, err := parseCreatePurchaseOrderToolInput(map[string]any{
		"supplier_id":            "42",
		"purchase_order_number":  " PO-1001 ",
		"tax":                    "11.5",
		"expected_delivery_date": "2026-07-01T10:30:00Z",
		"notes":                  " Restock protein bars ",
		"line_items": []any{
			map[string]any{"product_id": "10", "qty": "5", "cost": "12.25"},
			map[string]any{"product_id": 11, "quantity": 2.0, "price": 4},
		},
	})
	if err != nil {
		t.Fatalf("parseCreatePurchaseOrderToolInput: %v", err)
	}
	if parsed.SupplierID != 42 {
		t.Fatalf("SupplierID = %d", parsed.SupplierID)
	}
	if parsed.PONumber != "PO-1001" {
		t.Fatalf("PONumber = %q", parsed.PONumber)
	}
	if parsed.TaxAmount != 11.5 {
		t.Fatalf("TaxAmount = %v", parsed.TaxAmount)
	}
	if parsed.ExpectedDelivery != "2026-07-01" {
		t.Fatalf("ExpectedDelivery = %q", parsed.ExpectedDelivery)
	}
	if parsed.Notes != "Restock protein bars" {
		t.Fatalf("Notes = %q", parsed.Notes)
	}
	if len(parsed.Items) != 2 {
		t.Fatalf("len(Items) = %d", len(parsed.Items))
	}
	if parsed.Items[0].ProductID != 10 || parsed.Items[0].QuantityOrdered != 5 || parsed.Items[0].UnitCost != 12.25 {
		t.Fatalf("first item = %+v", parsed.Items[0])
	}
	if parsed.Items[1].ProductID != 11 || parsed.Items[1].QuantityOrdered != 2 || parsed.Items[1].UnitCost != 4 {
		t.Fatalf("second item = %+v", parsed.Items[1])
	}
}

func TestParseCreatePurchaseOrderToolInputAliases(t *testing.T) {
	parsed, err := parseCreatePurchaseOrderToolInput(map[string]any{
		"supplier_id": 5,
		"po":          "PO-ALIAS",
		"items": []map[string]any{
			{"product_id": 9, "units": 3, "unit_cost": 7.5},
		},
	})
	if err != nil {
		t.Fatalf("parseCreatePurchaseOrderToolInput: %v", err)
	}
	if parsed.PONumber != "PO-ALIAS" {
		t.Fatalf("PONumber = %q", parsed.PONumber)
	}
	if len(parsed.Items) != 1 || parsed.Items[0].QuantityOrdered != 3 {
		t.Fatalf("Items = %+v", parsed.Items)
	}
}

func TestParseCreatePurchaseOrderToolInputRejectsMissingSupplier(t *testing.T) {
	_, err := parseCreatePurchaseOrderToolInput(map[string]any{
		"po_number": "PO-1",
		"items":     []any{map[string]any{"product_id": 1, "quantity": 1, "unit_cost": 1}},
	})
	if err == nil {
		t.Fatal("expected missing supplier error")
	}
}

func TestParseCreatePurchaseOrderToolInputRejectsMissingPONumber(t *testing.T) {
	_, err := parseCreatePurchaseOrderToolInput(map[string]any{
		"supplier_id": 1,
		"items":       []any{map[string]any{"product_id": 1, "quantity": 1, "unit_cost": 1}},
	})
	if err == nil {
		t.Fatal("expected missing po number error")
	}
}

func TestParseCreatePurchaseOrderToolInputRejectsMissingItems(t *testing.T) {
	_, err := parseCreatePurchaseOrderToolInput(map[string]any{
		"supplier_id": 1,
		"po_number":   "PO-1",
	})
	if err == nil {
		t.Fatal("expected missing items error")
	}
}

func TestParseCreatePurchaseOrderToolInputRejectsInvalidExpectedDelivery(t *testing.T) {
	_, err := parseCreatePurchaseOrderToolInput(map[string]any{
		"supplier_id":       1,
		"po_number":         "PO-1",
		"expected_delivery": "tomorrow-ish",
		"items":             []any{map[string]any{"product_id": 1, "quantity": 1, "unit_cost": 1}},
	})
	if err == nil {
		t.Fatal("expected invalid expected delivery error")
	}
}

func TestParseCreatePurchaseOrderToolInputRejectsInvalidItem(t *testing.T) {
	_, err := parseCreatePurchaseOrderToolInput(map[string]any{
		"supplier_id": 1,
		"po_number":   "PO-1",
		"items":       []any{map[string]any{"product_id": 1, "quantity": 0, "unit_cost": 1}},
	})
	if err == nil {
		t.Fatal("expected invalid item error")
	}
}

func TestParseApprovePurchaseOrderToolInput(t *testing.T) {
	poID, err := parseApprovePurchaseOrderToolInput(map[string]any{"po_id": "55"})
	if err != nil {
		t.Fatalf("parseApprovePurchaseOrderToolInput: %v", err)
	}
	if poID != 55 {
		t.Fatalf("poID = %d", poID)
	}
}

func TestParseApprovePurchaseOrderToolInputRejectsMissingID(t *testing.T) {
	_, err := parseApprovePurchaseOrderToolInput(map[string]any{})
	if err == nil {
		t.Fatal("expected missing id error")
	}
}

func TestParsePurchaseOrderIDToolInputRejectsFractionalID(t *testing.T) {
	_, err := parsePurchaseOrderIDToolInput(map[string]any{"purchase_order_id": 4.5})
	if err == nil {
		t.Fatal("expected fractional id error")
	}
}
