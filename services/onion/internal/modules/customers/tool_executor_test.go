package customers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/customers/internal/domain"
)

func TestParseListCustomersToolInput(t *testing.T) {
	parsed, err := parseListCustomersToolInput(map[string]any{
		"query":        " ada ",
		"loyalty_tier": " gold ",
		"tag":          " vip ",
		"page":         json.Number("2"),
		"limit":        "10",
	})
	if err != nil {
		t.Fatalf("parseListCustomersToolInput: %v", err)
	}
	if parsed.Search != "ada" {
		t.Fatalf("Search = %q", parsed.Search)
	}
	if parsed.Tier != "gold" {
		t.Fatalf("Tier = %q", parsed.Tier)
	}
	if parsed.Tag != "vip" {
		t.Fatalf("Tag = %q", parsed.Tag)
	}
	if parsed.Page != 2 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.PageSize != 10 {
		t.Fatalf("PageSize = %d", parsed.PageSize)
	}
}

func TestParseListCustomersToolInputDefaultsAndCapsPageSize(t *testing.T) {
	parsed, err := parseListCustomersToolInput(map[string]any{
		"page_size": 500,
	})
	if err != nil {
		t.Fatalf("parseListCustomersToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.PageSize != maxListCustomersToolPageSize {
		t.Fatalf("PageSize = %d", parsed.PageSize)
	}
}

func TestParseListCustomersToolInputRejectsInvalidPage(t *testing.T) {
	_, err := parseListCustomersToolInput(map[string]any{
		"page": 1.5,
	})
	if err == nil {
		t.Fatal("expected invalid page error")
	}
}

func TestCustomerToolPayload(t *testing.T) {
	lastVisit := time.Date(2026, 6, 18, 10, 30, 0, 0, time.UTC)
	created := time.Date(2026, 6, 17, 9, 0, 0, 0, time.UTC)
	payload := customerToolPayload(domain.Customer{
		ID:            12,
		FirstName:     "Ada",
		LastName:      "Lovelace",
		Email:         "ada@example.com",
		Phone:         "+961000",
		LoyaltyPoints: 42,
		TotalSpent:    123.45,
		VisitCount:    3,
		LastVisitAt:   &lastVisit,
		Tags:          []string{"vip"},
		IsActive:      true,
		CreatedAt:     created,
	})
	if payload["full_name"] != "Ada Lovelace" {
		t.Fatalf("full_name = %#v", payload["full_name"])
	}
	if payload["last_visit_at"] != "2026-06-18T10:30:00Z" {
		t.Fatalf("last_visit_at = %#v", payload["last_visit_at"])
	}
	if payload["created_at"] != "2026-06-17T09:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
}
