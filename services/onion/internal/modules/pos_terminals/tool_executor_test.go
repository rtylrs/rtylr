package pos_terminals

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos_terminals/internal/domain"
)

func TestParseListPOSTerminalsToolInput(t *testing.T) {
	parsed, err := parseListPOSTerminalsToolInput(map[string]any{
		"search":      "front",
		"location_id": json.Number("4"),
		"status":      "active",
		"online":      "offline",
		"page":        2,
		"limit":       "10",
	})
	if err != nil {
		t.Fatalf("parseListPOSTerminalsToolInput: %v", err)
	}
	if parsed.Query != "front" {
		t.Fatalf("Query = %q", parsed.Query)
	}
	if parsed.BranchID != 4 {
		t.Fatalf("BranchID = %d", parsed.BranchID)
	}
	if parsed.Active == nil || !*parsed.Active {
		t.Fatalf("Active = %#v, want true", parsed.Active)
	}
	if parsed.Online == nil || *parsed.Online {
		t.Fatalf("Online = %#v, want false", parsed.Online)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListPOSTerminalsToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListPOSTerminalsToolInput(map[string]any{
		"limit": 500,
	})
	if err != nil {
		t.Fatalf("parseListPOSTerminalsToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListPOSTerminalsToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestParseListPOSTerminalsToolInputRejectsBadOnline(t *testing.T) {
	_, err := parseListPOSTerminalsToolInput(map[string]any{
		"online": "maybe",
	})
	if err == nil || !strings.Contains(err.Error(), "online") {
		t.Fatalf("error = %v, want online validation", err)
	}
}

func TestFilterPOSTerminalsForTool(t *testing.T) {
	active := true
	online := true
	lastSeenAt := time.Now().UTC().Add(-time.Minute)
	deviceID := "FRONT-1"
	terminals := []domain.Terminal{
		{ID: 1, BranchID: 4, Name: "Front Register", DeviceID: &deviceID, IsActive: true, LastSeenAt: &lastSeenAt},
		{ID: 2, BranchID: 4, Name: "Back Register", IsActive: true},
		{ID: 3, BranchID: 5, Name: "Front Register", IsActive: true, LastSeenAt: &lastSeenAt},
	}

	filtered := filterPOSTerminalsForTool(terminals, listPOSTerminalsToolInput{
		Query:    "front",
		BranchID: 4,
		Active:   &active,
		Online:   &online,
	})
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d", len(filtered))
	}
	if filtered[0].ID != 1 {
		t.Fatalf("filtered[0].ID = %d", filtered[0].ID)
	}
}

func TestPOSTerminalToolPayload(t *testing.T) {
	deviceID := "POS-1"
	lastSeenAt := time.Now().UTC().Add(-time.Minute)
	payload := posTerminalToolPayload(domain.Terminal{
		ID:         7,
		BusinessID: 2,
		BranchID:   4,
		Name:       "Front Register",
		DeviceID:   &deviceID,
		IsActive:   true,
		LastSeenAt: &lastSeenAt,
		CreatedAt:  lastSeenAt,
	})

	if payload["device_id"] != deviceID {
		t.Fatalf("device_id = %#v", payload["device_id"])
	}
	if payload["is_online"] != true {
		t.Fatalf("is_online = %#v", payload["is_online"])
	}
	if payload["last_seen_at"] != lastSeenAt.UTC().Format(time.RFC3339) {
		t.Fatalf("last_seen_at = %#v", payload["last_seen_at"])
	}
}
