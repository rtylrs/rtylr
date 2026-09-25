package pos_modifiers

import (
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos_modifiers/internal/domain"
)

func TestParseListPOSModifiersToolInput(t *testing.T) {
	parsed, err := parseListPOSModifiersToolInput(map[string]any{
		"search":     "cheese",
		"group_name": "extras",
		"required":   "optional",
		"page":       2,
		"limit":      "10",
	})
	if err != nil {
		t.Fatalf("parseListPOSModifiersToolInput: %v", err)
	}
	if parsed.Query != "cheese" {
		t.Fatalf("Query = %q", parsed.Query)
	}
	if parsed.Group != "extras" {
		t.Fatalf("Group = %q", parsed.Group)
	}
	if parsed.Required == nil || *parsed.Required {
		t.Fatalf("Required = %#v, want false", parsed.Required)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListPOSModifiersToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListPOSModifiersToolInput(map[string]any{
		"limit": 500,
	})
	if err != nil {
		t.Fatalf("parseListPOSModifiersToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListPOSModifiersToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestParseListPOSModifiersToolInputRejectsBadRequired(t *testing.T) {
	_, err := parseListPOSModifiersToolInput(map[string]any{
		"required": "sometimes",
	})
	if err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("error = %v, want required validation", err)
	}
}

func TestFilterPOSModifiersForTool(t *testing.T) {
	required := false
	modifiers := []domain.Modifier{
		{ID: 1, Name: "Extra Cheese", GroupName: "Extras", IsRequired: false},
		{ID: 2, Name: "Cheese Choice", GroupName: "Base", IsRequired: true},
		{ID: 3, Name: "Extra Sauce", GroupName: "Extras", IsRequired: false},
	}

	filtered := filterPOSModifiersForTool(modifiers, listPOSModifiersToolInput{
		Query:    "cheese",
		Group:    "extras",
		Required: &required,
	})
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d", len(filtered))
	}
	if filtered[0].ID != 1 {
		t.Fatalf("filtered[0].ID = %d", filtered[0].ID)
	}
}

func TestPOSModifierToolPayload(t *testing.T) {
	createdAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := posModifierToolPayload(domain.Modifier{
		ID:         7,
		OwnerID:    2,
		Name:       "Extra Cheese",
		PriceDelta: 1.5,
		GroupName:  "Extras",
		IsRequired: false,
		CreatedAt:  createdAt,
		UpdatedAt:  createdAt,
	})

	if payload["price_delta"] != 1.5 {
		t.Fatalf("price_delta = %#v", payload["price_delta"])
	}
	if payload["is_upcharge"] != true {
		t.Fatalf("is_upcharge = %#v", payload["is_upcharge"])
	}
	if payload["created_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
}
