package sessions

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos/sessions/internal/domain"
)

func TestParseListSessionsToolInput(t *testing.T) {
	parsed, err := parseListSessionsToolInput(map[string]any{
		"status":      "open",
		"location_id": json.Number("8"),
		"staff_id":    "12",
		"page":        2,
		"limit":       10,
	})
	if err != nil {
		t.Fatalf("parseListSessionsToolInput: %v", err)
	}
	if parsed.Status != "open" {
		t.Fatalf("Status = %q", parsed.Status)
	}
	if parsed.LocationID != 8 {
		t.Fatalf("LocationID = %d", parsed.LocationID)
	}
	if parsed.StaffID != 12 {
		t.Fatalf("StaffID = %d", parsed.StaffID)
	}
	if parsed.Page != 2 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != 10 {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestParseListSessionsToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListSessionsToolInput(map[string]any{
		"limit": 500,
	})
	if err != nil {
		t.Fatalf("parseListSessionsToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListSessionsToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestParseListSessionsToolInputRejectsBadStatus(t *testing.T) {
	_, err := parseListSessionsToolInput(map[string]any{
		"status": "settled",
	})
	if err == nil || !strings.Contains(err.Error(), "status") {
		t.Fatalf("error = %v, want status validation", err)
	}
}

func TestParseListSessionsToolInputRejectsOverflowingPage(t *testing.T) {
	_, err := parseListSessionsToolInput(map[string]any{
		"page": int64(500000000000000000),
	})
	if err == nil || !strings.Contains(err.Error(), "page") {
		t.Fatalf("error = %v, want page validation", err)
	}
}

func TestListSessionsToolBoundsNeverPanics(t *testing.T) {
	// (page-1)*limit overflows int64 for this input: the offset must clamp to
	// the end of the slice instead of going negative.
	start, end := listSessionsToolBounds(5, 500000000000000000, 20)
	if start != 5 || end != 5 {
		t.Fatalf("bounds = (%d, %d), want (5, 5)", start, end)
	}

	start, end = listSessionsToolBounds(5, 1, 20)
	if start != 0 || end != 5 {
		t.Fatalf("bounds = (%d, %d), want (0, 5)", start, end)
	}

	start, end = listSessionsToolBounds(5, 3, 2)
	if start != 4 || end != 5 {
		t.Fatalf("bounds = (%d, %d), want (4, 5)", start, end)
	}
}

func TestFilterSessionsForTool(t *testing.T) {
	locationID := int64(4)
	staffID := int64(7)
	sessions := []domain.Session{
		{ID: 1, Status: "open", LocationID: &locationID, StaffID: &staffID},
		{ID: 2, Status: "closed", LocationID: &locationID, StaffID: &staffID},
		{ID: 3, Status: "open"},
	}

	filtered := filterSessionsForTool(sessions, listSessionsToolInput{
		Status:     "open",
		LocationID: locationID,
		StaffID:    staffID,
	})
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d", len(filtered))
	}
	if filtered[0].ID != 1 {
		t.Fatalf("filtered[0].ID = %d", filtered[0].ID)
	}
}

func TestSessionToolPayload(t *testing.T) {
	locationID := int64(4)
	openedAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	closedAt := openedAt.Add(2 * time.Hour)
	payload := sessionToolPayload(domain.Session{
		ID:           5,
		OwnerID:      9,
		LocationID:   &locationID,
		OpeningFloat: 100,
		ClosingFloat: 250,
		TotalSales:   150,
		Status:       "closed",
		OpenedAt:     openedAt,
		ClosedAt:     &closedAt,
		CreatedAt:    openedAt,
	})

	if payload["location_id"] != locationID {
		t.Fatalf("location_id = %#v", payload["location_id"])
	}
	if payload["staff_id"] != nil {
		t.Fatalf("staff_id = %#v, want nil", payload["staff_id"])
	}
	if payload["opened_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("opened_at = %#v", payload["opened_at"])
	}
	if payload["closed_at"] != "2026-06-18T09:00:00Z" {
		t.Fatalf("closed_at = %#v", payload["closed_at"])
	}
}
