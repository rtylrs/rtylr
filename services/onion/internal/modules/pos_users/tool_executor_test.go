package pos_users

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos_users/internal/domain"
)

func TestParseListPOSUsersToolInput(t *testing.T) {
	parsed, err := parseListPOSUsersToolInput(map[string]any{
		"search":      "maya",
		"location_id": json.Number("4"),
		"role":        "cashier",
		"status":      "active",
		"page":        2,
		"limit":       "10",
	})
	if err != nil {
		t.Fatalf("parseListPOSUsersToolInput: %v", err)
	}
	if parsed.Query != "maya" {
		t.Fatalf("Query = %q", parsed.Query)
	}
	if parsed.LocationID != 4 {
		t.Fatalf("LocationID = %d", parsed.LocationID)
	}
	if parsed.Role != "cashier" {
		t.Fatalf("Role = %q", parsed.Role)
	}
	if parsed.Active == nil || !*parsed.Active {
		t.Fatalf("Active = %#v, want true", parsed.Active)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListPOSUsersToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListPOSUsersToolInput(map[string]any{
		"limit": 500,
	})
	if err != nil {
		t.Fatalf("parseListPOSUsersToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListPOSUsersToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestParseListPOSUsersToolInputRejectsBadRole(t *testing.T) {
	_, err := parseListPOSUsersToolInput(map[string]any{
		"role": "director",
	})
	if err == nil || !strings.Contains(err.Error(), "role") {
		t.Fatalf("error = %v, want role validation", err)
	}
}

func TestFilterPOSUsersForTool(t *testing.T) {
	active := true
	users := []domain.POSUser{
		{ID: 1, UserID: 11, Name: "Maya", Role: domain.RoleCashier, IsActive: true},
		{ID: 2, UserID: 12, Name: "Maya Old", Role: domain.RoleCashier, IsActive: false},
		{ID: 3, UserID: 13, Name: "Sam", Role: domain.RoleServer, IsActive: true},
	}

	filtered := filterPOSUsersForTool(users, listPOSUsersToolInput{
		Query:  "maya",
		Role:   "cashier",
		Active: &active,
	})
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d", len(filtered))
	}
	if filtered[0].ID != 1 {
		t.Fatalf("filtered[0].ID = %d", filtered[0].ID)
	}
}

func TestPOSUserToolPayloadOmitsSecrets(t *testing.T) {
	createdAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := posUserToolPayload(domain.POSUser{
		ID:               7,
		OwnerID:          2,
		UserID:           11,
		LocationID:       4,
		Name:             "Maya",
		Role:             domain.RoleCashier,
		IsActive:         true,
		HourlyRate:       42,
		MaxDiscountPct:   15,
		CanOpenDrawer:    true,
		CanApplyDiscount: true,
		CreatedAt:        createdAt,
		UpdatedAt:        createdAt,
	})

	if payload["role"] != "cashier" {
		t.Fatalf("role = %#v", payload["role"])
	}
	if payload["created_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
	for _, forbidden := range []string{"pin", "pin_hash", "hourly_rate"} {
		if _, ok := payload[forbidden]; ok {
			t.Fatalf("payload includes %q", forbidden)
		}
	}
}
