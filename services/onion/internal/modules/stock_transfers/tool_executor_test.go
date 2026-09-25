package stocktransfers

import (
	"strings"
	"testing"
	"time"

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

func TestParseStockTransferIDToolInputAcceptsAliases(t *testing.T) {
	cases := []map[string]any{
		{"stock_transfer_id": int64(55)},
		{"stock_transfer": "#55"},
		{"transfer": "transfer-55"},
		{"id": "stock-transfer-55"},
		{"id": "stock_transfer_55"},
		{"id": "st_55"},
	}

	for _, input := range cases {
		transferID, err := parseStockTransferIDToolInput(input)
		if err != nil {
			t.Fatalf("parseStockTransferIDToolInput(%v) returned error: %v", input, err)
		}
		if transferID != 55 {
			t.Fatalf("transferID = %d, want 55 for %v", transferID, input)
		}
	}
}

func TestParseStockTransferIDToolInputRejectsMissingID(t *testing.T) {
	_, err := parseStockTransferIDToolInput(map[string]any{})
	if err == nil {
		t.Fatal("expected missing id error")
	}
}

func TestParseListStockTransfersToolInput(t *testing.T) {
	parsed, err := parseListStockTransfersToolInput(map[string]any{
		"status":    "in transit",
		"page":      "2",
		"page_size": "500",
	})
	if err != nil {
		t.Fatalf("parseListStockTransfersToolInput: %v", err)
	}
	if parsed.Status != domain.StockTransferStatusInTransit {
		t.Fatalf("Status = %q, want in_transit", parsed.Status)
	}
	if parsed.Page != 2 {
		t.Fatalf("Page = %d, want 2", parsed.Page)
	}
	if parsed.PageSize != maxListStockTransfersToolPageSize {
		t.Fatalf("PageSize = %d, want %d", parsed.PageSize, maxListStockTransfersToolPageSize)
	}
}

func TestParseListStockTransfersToolInputRejectsBadStatus(t *testing.T) {
	_, err := parseListStockTransfersToolInput(map[string]any{"status": "pending"})
	if err == nil || !strings.Contains(err.Error(), "status") {
		t.Fatalf("error = %v, want status validation", err)
	}
}

func TestStockTransferToolPayloadFormatsFields(t *testing.T) {
	createdAt := time.Date(2026, 6, 18, 11, 30, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := stockTransferToolPayload(domain.StockTransfer{
		ID:              44,
		FromWarehouseID: 3,
		ToWarehouseID:   4,
		Status:          domain.StockTransferStatusDraft,
		Notes:           "Move summer stock",
		CreatedAt:       createdAt,
		Items: []domain.StockTransferItem{
			{ID: 1, TransferID: 44, ProductID: 10, Quantity: 5.5, CreatedAt: createdAt},
		},
	})

	if payload["id"] != int64(44) {
		t.Fatalf("id = %v", payload["id"])
	}
	if payload["status"] != "draft" {
		t.Fatalf("status = %v", payload["status"])
	}
	if payload["item_count"] != 1 {
		t.Fatalf("item_count = %v", payload["item_count"])
	}
	if payload["created_at"] != "2026-06-18T08:30:00Z" {
		t.Fatalf("created_at = %v", payload["created_at"])
	}
}
