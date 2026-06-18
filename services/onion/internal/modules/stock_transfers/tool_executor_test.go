package stocktransfers

import (
	"testing"

	"github.com/voxire/onion/internal/modules/stock_transfers/internal/domain"
)

func TestParseCreateStockTransferToolInput(t *testing.T) {
	parsed, err := parseCreateStockTransferToolInput(map[string]any{
		"source_warehouse_id":      "3",
		"destination_warehouse_id": 4,
		"notes":                    " Move summer stock ",
		"line_items": []any{
			map[string]any{"product_id": "10", "qty": "5.5"},
			map[string]any{"product_id": 11, "units": 2.0},
		},
	})
	if err != nil {
		t.Fatalf("parseCreateStockTransferToolInput: %v", err)
	}
	if parsed.FromWarehouseID != 3 || parsed.ToWarehouseID != 4 {
		t.Fatalf("warehouse ids = from %d to %d", parsed.FromWarehouseID, parsed.ToWarehouseID)
	}
	if parsed.Notes != "Move summer stock" {
		t.Fatalf("Notes = %q", parsed.Notes)
	}
	if len(parsed.Items) != 2 {
		t.Fatalf("len(Items) = %d", len(parsed.Items))
	}
	if parsed.Items[0].ProductID != 10 || parsed.Items[0].Quantity != 5.5 {
		t.Fatalf("first item = %+v", parsed.Items[0])
	}
	if parsed.Items[1].ProductID != 11 || parsed.Items[1].Quantity != 2 {
		t.Fatalf("second item = %+v", parsed.Items[1])
	}
}

func TestParseCreateStockTransferToolInputRejectsSameWarehouse(t *testing.T) {
	_, err := parseCreateStockTransferToolInput(map[string]any{
		"from_warehouse_id": 1,
		"to_warehouse_id":   1,
		"items":             []any{map[string]any{"product_id": 1, "quantity": 1}},
	})
	if err == nil {
		t.Fatal("expected same warehouse error")
	}
}

func TestParseCreateStockTransferToolInputRejectsMissingItems(t *testing.T) {
	_, err := parseCreateStockTransferToolInput(map[string]any{
		"from_warehouse_id": 1,
		"to_warehouse_id":   2,
	})
	if err == nil {
		t.Fatal("expected missing items error")
	}
}

func TestParseCreateStockTransferToolInputRejectsInvalidItem(t *testing.T) {
	_, err := parseCreateStockTransferToolInput(map[string]any{
		"from_warehouse_id": 1,
		"to_warehouse_id":   2,
		"items":             []any{map[string]any{"product_id": 1, "quantity": -1}},
	})
	if err == nil {
		t.Fatal("expected invalid item error")
	}
}

func TestParseCreateStockTransferToolInputRejectsInvalidTypedItem(t *testing.T) {
	_, err := parseCreateStockTransferToolInput(map[string]any{
		"from_warehouse_id": 1,
		"to_warehouse_id":   2,
		"items":             []domain.StockTransferItemInput{{ProductID: 1, Quantity: -1}},
	})
	if err == nil {
		t.Fatal("expected invalid typed item error")
	}
}

func TestParseStockTransferIDToolInput(t *testing.T) {
	transferID, err := parseStockTransferIDToolInput(map[string]any{"transfer_id": "55"})
	if err != nil {
		t.Fatalf("parseStockTransferIDToolInput: %v", err)
	}
	if transferID != 55 {
		t.Fatalf("transferID = %d", transferID)
	}
}

func TestParseStockTransferIDToolInputRejectsMissingID(t *testing.T) {
	_, err := parseStockTransferIDToolInput(map[string]any{})
	if err == nil {
		t.Fatal("expected missing id error")
	}
}
