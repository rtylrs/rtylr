package floors

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos/floors/internal/domain"
)

func TestParseListPOSFloorsToolInput(t *testing.T) {
	parsed, err := parseListPOSFloorsToolInput(map[string]any{
		"location_id": json.Number("4"),
		"search":      "main",
		"active":      1.0,
		"page":        2,
		"limit":       "10",
	})
	if err != nil {
		t.Fatalf("parseListPOSFloorsToolInput: %v", err)
	}
	if parsed.BranchID != 4 {
		t.Fatalf("BranchID = %d", parsed.BranchID)
	}
	if parsed.Query != "main" {
		t.Fatalf("Query = %q", parsed.Query)
	}
	if parsed.Active == nil || !*parsed.Active {
		t.Fatalf("Active = %#v, want true", parsed.Active)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListPOSFloorsToolInputRequiresBranchID(t *testing.T) {
	_, err := parseListPOSFloorsToolInput(map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "branch_id") {
		t.Fatalf("error = %v, want branch_id validation", err)
	}
}

func TestParseListPOSFloorsToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListPOSFloorsToolInput(map[string]any{
		"branch_id": 4,
		"limit":     500,
	})
	if err != nil {
		t.Fatalf("parseListPOSFloorsToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListPOSFloorsToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestFilterPOSFloorsForTool(t *testing.T) {
	active := true
	floors := []domain.Floor{
		{ID: 1, Name: "Main Floor", IsActive: true},
		{ID: 2, Name: "Main Old", IsActive: false},
		{ID: 3, Name: "Patio", IsActive: true},
	}

	filtered := filterPOSFloorsForTool(floors, listPOSFloorsToolInput{
		Query:  "main",
		Active: &active,
	})
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d", len(filtered))
	}
	if filtered[0].ID != 1 {
		t.Fatalf("filtered[0].ID = %d", filtered[0].ID)
	}
}

func TestPOSFloorToolPayload(t *testing.T) {
	createdAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := posFloorToolPayload(domain.Floor{
		ID:         7,
		BusinessID: 2,
		BranchID:   4,
		Name:       "Main Floor",
		SortOrder:  3,
		IsActive:   true,
		CreatedAt:  createdAt,
	})

	if payload["sort_order"] != 3 {
		t.Fatalf("sort_order = %#v", payload["sort_order"])
	}
	if payload["created_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
}
