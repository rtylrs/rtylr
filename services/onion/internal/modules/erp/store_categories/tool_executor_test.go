package store_categories

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/erp/store_categories/internal/domain"
)

func TestParseListStoreCategoriesToolInput(t *testing.T) {
	parsed, err := parseListStoreCategoriesToolInput(map[string]any{
		"search":    "protein",
		"parent_id": json.Number("4"),
		"page":      2,
		"limit":     "10",
	})
	if err != nil {
		t.Fatalf("parseListStoreCategoriesToolInput: %v", err)
	}
	if parsed.Query != "protein" {
		t.Fatalf("Query = %q", parsed.Query)
	}
	if parsed.ParentID != 4 {
		t.Fatalf("ParentID = %d", parsed.ParentID)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListStoreCategoriesToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListStoreCategoriesToolInput(map[string]any{
		"limit": 500,
	})
	if err != nil {
		t.Fatalf("parseListStoreCategoriesToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListStoreCategoriesToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestParseListStoreCategoriesToolInputRejectsNegativeParentID(t *testing.T) {
	_, err := parseListStoreCategoriesToolInput(map[string]any{
		"parent_id": -1,
	})
	if err == nil || !strings.Contains(err.Error(), "parent_id") {
		t.Fatalf("error = %v, want parent_id validation", err)
	}
}

func TestFilterStoreCategoriesForTool(t *testing.T) {
	categories := []domain.StoreCategory{
		{ID: 1, Name: "Protein", Slug: "protein", ParentID: 4},
		{ID: 2, Name: "Protein Bars", Slug: "protein-bars", ParentID: 5},
		{ID: 3, Name: "Creatine", Slug: "creatine", ParentID: 4},
	}

	filtered := filterStoreCategoriesForTool(categories, listStoreCategoriesToolInput{
		Query:    "protein",
		ParentID: 4,
	})
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d", len(filtered))
	}
	if filtered[0].ID != 1 {
		t.Fatalf("filtered[0].ID = %d", filtered[0].ID)
	}
}

func TestStoreCategoryToolPayload(t *testing.T) {
	createdAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := storeCategoryToolPayload(domain.StoreCategory{
		ID:        7,
		OwnerID:   2,
		Name:      "Protein",
		Slug:      "protein",
		ParentID:  4,
		CreatedAt: createdAt,
	})

	if payload["parent_id"] != int64(4) {
		t.Fatalf("parent_id = %#v", payload["parent_id"])
	}
	if payload["created_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
}
