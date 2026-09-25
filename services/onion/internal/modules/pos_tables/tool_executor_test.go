package pos_tables

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos_tables/internal/domain"
)

func TestParseListPOSTablesToolInput(t *testing.T) {
	parsed, err := parseListPOSTablesToolInput(map[string]any{
		"label":       "patio",
		"location_id": json.Number("4"),
		"status":      "reserved",
		"min_seats":   4.0,
		"page":        2,
		"limit":       "10",
	})
	if err != nil {
		t.Fatalf("parseListPOSTablesToolInput: %v", err)
	}
	if parsed.Query != "patio" {
		t.Fatalf("Query = %q", parsed.Query)
	}
	if parsed.LocationID != 4 {
		t.Fatalf("LocationID = %d", parsed.LocationID)
	}
	if parsed.Status != "reserved" {
		t.Fatalf("Status = %q", parsed.Status)
	}
	if parsed.MinSeats != 4 {
		t.Fatalf("MinSeats = %d", parsed.MinSeats)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListPOSTablesToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListPOSTablesToolInput(map[string]any{
		"limit": 500,
	})
	if err != nil {
		t.Fatalf("parseListPOSTablesToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListPOSTablesToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestParseListPOSTablesToolInputRejectsBadStatus(t *testing.T) {
	_, err := parseListPOSTablesToolInput(map[string]any{
		"status": "broken",
	})
	if err == nil || !strings.Contains(err.Error(), "status") {
		t.Fatalf("error = %v, want status validation", err)
	}
}

func TestFilterPOSTablesForTool(t *testing.T) {
	locationID := int64(4)
	otherLocationID := int64(5)
	tables := []domain.Table{
		{ID: 1, LocationID: &locationID, Label: "Patio 1", Seats: 4, Status: "reserved"},
		{ID: 2, LocationID: &locationID, Label: "Patio 2", Seats: 2, Status: "reserved"},
		{ID: 3, LocationID: &otherLocationID, Label: "Patio 3", Seats: 4, Status: "reserved"},
		{ID: 4, LocationID: &locationID, Label: "Main 1", Seats: 4, Status: "available"},
	}

	filtered := filterPOSTablesForTool(tables, listPOSTablesToolInput{
		Query:      "patio",
		LocationID: locationID,
		Status:     "reserved",
		MinSeats:   4,
	})
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d", len(filtered))
	}
	if filtered[0].ID != 1 {
		t.Fatalf("filtered[0].ID = %d", filtered[0].ID)
	}
}

func TestPOSTableToolPayload(t *testing.T) {
	locationID := int64(4)
	createdAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := posTableToolPayload(domain.Table{
		ID:         7,
		OwnerID:    2,
		LocationID: &locationID,
		Label:      "Patio 1",
		Seats:      4,
		Status:     "available",
		CreatedAt:  createdAt,
		UpdatedAt:  createdAt,
	})

	if payload["location_id"] != locationID {
		t.Fatalf("location_id = %#v", payload["location_id"])
	}
	if payload["is_available"] != true {
		t.Fatalf("is_available = %#v", payload["is_available"])
	}
	if payload["created_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
}
