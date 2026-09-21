package pos_stock_movements

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos_stock_movements/internal/domain"
)

func TestParseListStockMovementsToolInput(t *testing.T) {
	parsed, err := parseListStockMovementsToolInput(map[string]any{
		"branch_id":   json.Number("4"),
		"product_id":  "10",
		"type":        "transfer out",
		"employee_id": 7,
		"page":        2,
		"limit":       10,
	})
	if err != nil {
		t.Fatalf("parseListStockMovementsToolInput: %v", err)
	}
	if parsed.BranchID != 4 || parsed.ProductID != 10 {
		t.Fatalf("ids = branch:%d product:%d", parsed.BranchID, parsed.ProductID)
	}
	if parsed.Type != "transfer_out" {
		t.Fatalf("Type = %q", parsed.Type)
	}
	if parsed.EmployeeID != 7 {
		t.Fatalf("EmployeeID = %d", parsed.EmployeeID)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListStockMovementsToolInputRequiresIDs(t *testing.T) {
	_, err := parseListStockMovementsToolInput(map[string]any{
		"branch_id": 4,
	})
	if err == nil || !strings.Contains(err.Error(), "product_id") {
		t.Fatalf("error = %v, want product_id validation", err)
	}
}

func TestParseListStockMovementsToolInputRejectsBadType(t *testing.T) {
	_, err := parseListStockMovementsToolInput(map[string]any{
		"branch_id":  4,
		"product_id": 10,
		"type":       "mystery",
	})
	if err == nil || !strings.Contains(err.Error(), "type") {
		t.Fatalf("error = %v, want type validation", err)
	}
}

func TestParseListBranchInventoryToolInputDefaultsAndCaps(t *testing.T) {
	parsed, err := parseListBranchInventoryToolInput(map[string]any{
		"branch_id": 4,
		"limit":     500,
	})
	if err != nil {
		t.Fatalf("parseListBranchInventoryToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxPOSStockToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestFilterStockMovementsForTool(t *testing.T) {
	employeeID := int64(7)
	movements := []domain.StockMovement{
		{ID: 1, Type: "receive", EmployeeID: &employeeID},
		{ID: 2, Type: "waste", EmployeeID: &employeeID},
		{ID: 3, Type: "receive"},
	}

	filtered := filterStockMovementsForTool(movements, listStockMovementsToolInput{
		Type:       "receive",
		EmployeeID: employeeID,
	})
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d", len(filtered))
	}
	if filtered[0].ID != 1 {
		t.Fatalf("filtered[0].ID = %d", filtered[0].ID)
	}
}

func TestStockMovementToolPayload(t *testing.T) {
	referenceID := int64(22)
	referenceType := "purchase_order"
	note := "Received from supplier"
	employeeID := int64(7)
	createdAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := stockMovementToolPayload(domain.StockMovement{
		ID:            5,
		BusinessID:    2,
		BranchID:      4,
		ProductID:     10,
		Type:          "receive",
		Quantity:      12,
		ReferenceID:   &referenceID,
		ReferenceType: &referenceType,
		Note:          &note,
		EmployeeID:    &employeeID,
		CreatedAt:     createdAt,
	})

	if payload["reference_id"] != referenceID {
		t.Fatalf("reference_id = %#v", payload["reference_id"])
	}
	if payload["reference_type"] != referenceType {
		t.Fatalf("reference_type = %#v", payload["reference_type"])
	}
	if payload["created_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
}

func TestBranchInventoryToolPayload(t *testing.T) {
	threshold := 5.0
	updatedAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := branchInventoryToolPayload(domain.BranchInventory{
		ID:                9,
		BusinessID:        2,
		BranchID:          4,
		ProductID:         10,
		Name:              "Whey",
		SKU:               "WHEY-1",
		Unit:              "unit",
		Quantity:          3,
		LowStockThreshold: &threshold,
		UpdatedAt:         updatedAt,
	})

	if payload["item_id"] != int64(10) {
		t.Fatalf("item_id = %#v", payload["item_id"])
	}
	if payload["low_stock_threshold"] != threshold {
		t.Fatalf("low_stock_threshold = %#v", payload["low_stock_threshold"])
	}
	if payload["updated_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("updated_at = %#v", payload["updated_at"])
	}
}

func TestPaginateStockMovementsForTool(t *testing.T) {
	items := []domain.StockMovement{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5}}

	page, hasMore := paginateStockMovementsForTool(items, 2, 2)
	if len(page) != 2 || page[0].ID != 3 || !hasMore {
		t.Fatalf("page 2 = %+v hasMore=%v", page, hasMore)
	}

	page, hasMore = paginateStockMovementsForTool(items, 3, 2)
	if len(page) != 1 || page[0].ID != 5 || hasMore {
		t.Fatalf("page 3 = %+v hasMore=%v", page, hasMore)
	}

	page, hasMore = paginateStockMovementsForTool(items, 100, 2)
	if len(page) != 0 || hasMore {
		t.Fatalf("past-the-end page = %+v hasMore=%v", page, hasMore)
	}
}

func TestPaginateStockMovementsForToolHugePageDoesNotPanic(t *testing.T) {
	items := []domain.StockMovement{{ID: 1}, {ID: 2}, {ID: 3}}

	// JSON decoding into map[string]any produces float64, so page 2^62 is
	// exactly representable and passes parsePOSStockToolPage.
	parsed, err := parseListStockMovementsToolInput(map[string]any{
		"branch_id":  4,
		"product_id": 10,
		"page":       float64(int64(1) << 62),
	})
	if err != nil {
		t.Fatalf("parseListStockMovementsToolInput: %v", err)
	}

	page, hasMore := paginateStockMovementsForTool(items, parsed.Page, parsed.Limit)
	if len(page) != 0 {
		t.Fatalf("len(page) = %d, want 0", len(page))
	}
	if hasMore {
		t.Fatalf("hasMore = true, want false")
	}
}

func TestPaginateBranchInventoryForToolHugePageDoesNotPanic(t *testing.T) {
	items := []domain.BranchInventory{{ID: 1}, {ID: 2}, {ID: 3}}

	page, hasMore := paginateBranchInventoryForTool(items, int64(1)<<62, defaultPOSStockToolLimit)
	if len(page) != 0 {
		t.Fatalf("len(page) = %d, want 0", len(page))
	}
	if hasMore {
		t.Fatalf("hasMore = true, want false")
	}
}
