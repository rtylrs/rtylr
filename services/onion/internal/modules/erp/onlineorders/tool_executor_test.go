package onlineorders

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/erp/onlineorders/internal/domain"
)

func TestParseListStoreOrdersToolInput(t *testing.T) {
	parsed, err := parseListStoreOrdersToolInput(map[string]any{
		"customer":  "maya",
		"status":    "paid",
		"min_total": json.Number("50.5"),
		"page":      2,
		"limit":     "10",
	})
	if err != nil {
		t.Fatalf("parseListStoreOrdersToolInput: %v", err)
	}
	if parsed.Query != "maya" {
		t.Fatalf("Query = %q", parsed.Query)
	}
	if parsed.Status != "paid" {
		t.Fatalf("Status = %q", parsed.Status)
	}
	if parsed.MinTotal != 50.5 {
		t.Fatalf("MinTotal = %g", parsed.MinTotal)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListStoreOrdersToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListStoreOrdersToolInput(map[string]any{
		"limit": 500,
	})
	if err != nil {
		t.Fatalf("parseListStoreOrdersToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListStoreOrdersToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestParseListStoreOrdersToolInputRejectsNegativeMinTotal(t *testing.T) {
	_, err := parseListStoreOrdersToolInput(map[string]any{
		"min_total": -1,
	})
	if err == nil || !strings.Contains(err.Error(), "min_total") {
		t.Fatalf("error = %v, want min_total validation", err)
	}
}

func TestFilterStoreOrdersForTool(t *testing.T) {
	orders := []domain.StoreOrder{
		{ID: 1, CustomerName: "Maya", Status: "paid", Total: 75},
		{ID: 2, CustomerName: "Maya", Status: "pending", Total: 75},
		{ID: 3, CustomerName: "Maya", Status: "paid", Total: 25},
		{ID: 4, CustomerName: "Sam", Status: "paid", Total: 75},
	}

	filtered := filterStoreOrdersForTool(orders, listStoreOrdersToolInput{
		Query:    "maya",
		Status:   "paid",
		MinTotal: 50,
	})
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d", len(filtered))
	}
	if filtered[0].ID != 1 {
		t.Fatalf("filtered[0].ID = %d", filtered[0].ID)
	}
}

func TestStoreOrderToolPayload(t *testing.T) {
	createdAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := storeOrderToolPayload(domain.StoreOrder{
		ID:            7,
		OwnerID:       2,
		CustomerEmail: "maya@example.com",
		Status:        "paid",
		Total:         75,
		CreatedAt:     createdAt,
		UpdatedAt:     createdAt,
	})

	if payload["total"] != float64(75) {
		t.Fatalf("total = %#v", payload["total"])
	}
	if payload["created_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
}

func TestStoreOrderItemToolPayload(t *testing.T) {
	createdAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := storeOrderItemToolPayload(domain.StoreOrderItem{
		ID:        9,
		OrderID:   7,
		ProductID: 3,
		Name:      "Whey",
		Quantity:  2,
		Price:     10,
		Total:     20,
		CreatedAt: createdAt,
	})

	if payload["quantity"] != 2 {
		t.Fatalf("quantity = %#v", payload["quantity"])
	}
	if payload["created_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
}
