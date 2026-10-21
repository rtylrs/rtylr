package supportcustomers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/support/customers/internal/domain"
)

func TestParseListSupportCustomersToolInput(t *testing.T) {
	parsed, err := parseListSupportCustomersToolInput(map[string]any{
		"page":      json.Number("2"),
		"page_size": "10",
	})
	if err != nil {
		t.Fatalf("parseListSupportCustomersToolInput: %v", err)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListSupportCustomersToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListSupportCustomersToolInput(map[string]any{
		"limit": 500,
	})
	if err != nil {
		t.Fatalf("parseListSupportCustomersToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListSupportCustomersToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestSupportCustomerToolPayload(t *testing.T) {
	createdAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	lastTicketAt := createdAt.Add(time.Hour)
	payload := supportCustomerToolPayload(domain.Customer{
		ID:           7,
		BusinessID:   2,
		Name:         "Maya",
		Email:        "maya@example.com",
		TotalTickets: 3,
		LastTicketAt: &lastTicketAt,
		CreatedAt:    createdAt,
	})

	if payload["total_tickets"] != 3 {
		t.Fatalf("total_tickets = %#v", payload["total_tickets"])
	}
	if payload["last_ticket_at"] != "2026-06-18T08:00:00Z" {
		t.Fatalf("last_ticket_at = %#v", payload["last_ticket_at"])
	}
	if payload["created_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
}
