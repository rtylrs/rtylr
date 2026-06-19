package supporttickets

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/support_tickets/internal/domain"
)

func TestParseListSupportTicketsToolInput(t *testing.T) {
	parsed, err := parseListSupportTicketsToolInput(map[string]any{
		"status":    "resolved",
		"page":      json.Number("2"),
		"page_size": "10",
	})
	if err != nil {
		t.Fatalf("parseListSupportTicketsToolInput: %v", err)
	}
	if parsed.Status != "resolved" {
		t.Fatalf("Status = %q", parsed.Status)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListSupportTicketsToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListSupportTicketsToolInput(map[string]any{
		"status": "all",
		"limit":  500,
	})
	if err != nil {
		t.Fatalf("parseListSupportTicketsToolInput: %v", err)
	}
	if parsed.Status != "" {
		t.Fatalf("Status = %q", parsed.Status)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListSupportTicketsToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestParseListSupportTicketsToolInputRejectsBadStatus(t *testing.T) {
	_, err := parseListSupportTicketsToolInput(map[string]any{
		"status": "waiting",
	})
	if err == nil || !strings.Contains(err.Error(), "status") {
		t.Fatalf("error = %v, want status validation", err)
	}
}

func TestSupportTicketToolPayload(t *testing.T) {
	assignedTo := int64(9)
	createdAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := supportTicketToolPayload(domain.Ticket{
		ID:            7,
		BusinessID:    2,
		CustomerName:  "Maya",
		CustomerEmail: "maya@example.com",
		Subject:       "Where is my order?",
		Status:        "open",
		Priority:      "urgent",
		AssignedTo:    &assignedTo,
		Channel:       "email",
		CreatedAt:     createdAt,
		UpdatedAt:     createdAt,
	})

	if payload["assigned_to"] != int64(9) {
		t.Fatalf("assigned_to = %#v", payload["assigned_to"])
	}
	if payload["created_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
}

func TestSupportMessageToolPayload(t *testing.T) {
	createdAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := supportMessageToolPayload(domain.Message{
		ID:         11,
		TicketID:   7,
		SenderType: "customer",
		SenderName: "Maya",
		Body:       "hello",
		CreatedAt:  createdAt,
	})

	if payload["body"] != "hello" {
		t.Fatalf("body = %#v", payload["body"])
	}
	if payload["created_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
}
