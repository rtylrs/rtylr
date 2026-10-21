package catalog

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/erp/products/internal/catalog/internal/domain"
)

func TestParseListPOSMenuItemsToolInput(t *testing.T) {
	parsed, err := parseListPOSMenuItemsToolInput(map[string]any{
		"search":      "burger",
		"category_id": json.Number("4"),
		"category":    "food",
		"available":   true,
		"page":        2,
		"limit":       "10",
	})
	if err != nil {
		t.Fatalf("parseListPOSMenuItemsToolInput: %v", err)
	}
	if parsed.Query != "burger" {
		t.Fatalf("Query = %q", parsed.Query)
	}
	if parsed.CategoryID != 4 {
		t.Fatalf("CategoryID = %d", parsed.CategoryID)
	}
	if parsed.Category != "food" {
		t.Fatalf("Category = %q", parsed.Category)
	}
	if parsed.Active == nil || !*parsed.Active {
		t.Fatalf("Active = %#v, want true", parsed.Active)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListPOSMenuItemsToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListPOSMenuItemsToolInput(map[string]any{
		"limit": 500,
	})
	if err != nil {
		t.Fatalf("parseListPOSMenuItemsToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListPOSMenuItemsToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestParseListPOSMenuItemsToolInputRejectsBadActive(t *testing.T) {
	_, err := parseListPOSMenuItemsToolInput(map[string]any{
		"active": "maybe",
	})
	if err == nil || !strings.Contains(err.Error(), "active") {
		t.Fatalf("error = %v, want active validation", err)
	}
}

func TestFilterPOSMenuItemsForTool(t *testing.T) {
	active := true
	items := []domain.MenuItem{
		{ID: 1, Name: "Burger", CategoryID: 4, Category: "Food", IsActive: true},
		{ID: 2, Name: "Burger Old", CategoryID: 4, Category: "Food", IsActive: false},
		{ID: 3, Name: "Burger Drink", CategoryID: 5, Category: "Drinks", IsActive: true},
	}

	filtered := filterPOSMenuItemsForTool(items, listPOSMenuItemsToolInput{
		Query:      "burger",
		CategoryID: 4,
		Category:   "food",
		Active:     &active,
	})
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d", len(filtered))
	}
	if filtered[0].ID != 1 {
		t.Fatalf("filtered[0].ID = %d", filtered[0].ID)
	}
}

func TestPOSMenuItemToolPayload(t *testing.T) {
	createdAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := posMenuItemToolPayload(domain.MenuItem{
		ID:          7,
		OwnerID:     2,
		CategoryID:  4,
		Category:    "Food",
		Name:        "Burger",
		SKU:         "BURGER",
		SalePrice:   12.5,
		Description: "Classic",
		IsActive:    true,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	})

	if payload["sale_price"] != 12.5 {
		t.Fatalf("sale_price = %#v", payload["sale_price"])
	}
	if payload["available"] != true {
		t.Fatalf("available = %#v", payload["available"])
	}
	if payload["created_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
}
