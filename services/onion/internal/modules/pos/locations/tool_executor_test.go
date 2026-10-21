package locations

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos/locations/internal/domain"
)

func TestParseListPOSLocationsToolInput(t *testing.T) {
	parsed, err := parseListPOSLocationsToolInput(map[string]any{
		"q":      "Downtown",
		"status": "active",
		"page":   json.Number("2"),
		"limit":  "10",
	})
	if err != nil {
		t.Fatalf("parseListPOSLocationsToolInput: %v", err)
	}
	if parsed.Query != "Downtown" {
		t.Fatalf("Query = %q", parsed.Query)
	}
	if parsed.Active == nil || !*parsed.Active {
		t.Fatalf("Active = %#v, want true", parsed.Active)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListPOSLocationsToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListPOSLocationsToolInput(map[string]any{
		"limit": 500,
	})
	if err != nil {
		t.Fatalf("parseListPOSLocationsToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListPOSLocationsToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestParseListPOSLocationsToolInputRejectsBadActive(t *testing.T) {
	_, err := parseListPOSLocationsToolInput(map[string]any{
		"active": "sometimes",
	})
	if err == nil || !strings.Contains(err.Error(), "active") {
		t.Fatalf("error = %v, want active validation", err)
	}
}

func TestFilterPOSLocationsForTool(t *testing.T) {
	active := true
	locations := []domain.Location{
		{ID: 1, Name: "Downtown", Address: "Main Street", IsActive: true},
		{ID: 2, Name: "Warehouse", Address: "Industrial Zone", IsActive: true},
		{ID: 3, Name: "Downtown Old", Address: "Old Street", IsActive: false},
	}

	filtered := filterPOSLocationsForTool(locations, listPOSLocationsToolInput{
		Query:  "down",
		Active: &active,
	})
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d", len(filtered))
	}
	if filtered[0].ID != 1 {
		t.Fatalf("filtered[0].ID = %d", filtered[0].ID)
	}
}

func TestPOSLocationToolPayload(t *testing.T) {
	createdAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := posLocationToolPayload(domain.Location{
		ID:        7,
		OwnerID:   2,
		Name:      "Downtown",
		Address:   "Main Street",
		Phone:     "+961",
		IsActive:  true,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	})

	if payload["id"] != int64(7) {
		t.Fatalf("id = %#v", payload["id"])
	}
	if payload["created_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
}
