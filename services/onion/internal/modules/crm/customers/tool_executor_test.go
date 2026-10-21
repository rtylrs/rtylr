package customers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
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

func TestParseGetCustomerToolInputAcceptsAliases(t *testing.T) {
	id, err := parseGetCustomerToolInput(map[string]any{
		"id": json.Number("42"),
	})
	if err != nil {
		t.Fatalf("parseGetCustomerToolInput: %v", err)
	}
	if id != 42 {
		t.Fatalf("id = %d, want 42", id)
	}

	id, err = parseGetCustomerToolInput(map[string]any{
		"customer_id": "77",
	})
	if err != nil {
		t.Fatalf("parseGetCustomerToolInput customer_id: %v", err)
	}
	if id != 77 {
		t.Fatalf("id = %d, want 77", id)
	}
}

func TestParseGetCustomerToolInputRequiresCustomerID(t *testing.T) {
	_, err := parseGetCustomerToolInput(map[string]any{})
	if err == nil {
		t.Fatal("expected required customer_id error")
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

func TestCustomerDetailToolPayloadIncludesProfileFields(t *testing.T) {
	birthDate := time.Date(1990, 2, 3, 0, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	payload := customerDetailToolPayload(domain.Customer{
		ID:          12,
		BusinessID:  99,
		FirstName:   "Ada",
		LastName:    "Lovelace",
		DateOfBirth: &birthDate,
		Gender:      "female",
		Address:     "12 Market St",
		City:        "Beirut",
		Country:     "Lebanon",
		Notes:       "Prefers SMS",
		UpdatedAt:   updated,
	})

	if payload["business_id"] != int64(99) {
		t.Fatalf("business_id = %#v", payload["business_id"])
	}
	if payload["date_of_birth"] != "1990-02-03" {
		t.Fatalf("date_of_birth = %#v", payload["date_of_birth"])
	}
	if payload["notes"] != "Prefers SMS" || payload["city"] != "Beirut" {
		t.Fatalf("missing detail fields: %#v", payload)
	}
	if payload["updated_at"] != "2026-06-19T12:00:00Z" {
		t.Fatalf("updated_at = %#v", payload["updated_at"])
	}
}
