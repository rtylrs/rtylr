package supportinbox

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/support/inbox/internal/domain"
)

func TestParseListSupportInboxToolInput(t *testing.T) {
	parsed, err := parseListSupportInboxToolInput(map[string]any{
		"page":      json.Number("2"),
		"page_size": "10",
	})
	if err != nil {
		t.Fatalf("parseListSupportInboxToolInput: %v", err)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListSupportInboxToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListSupportInboxToolInput(map[string]any{
		"limit": 500,
	})
	if err != nil {
		t.Fatalf("parseListSupportInboxToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListSupportInboxToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestSupportInboxItemToolPayload(t *testing.T) {
	lastMessageAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := supportInboxItemToolPayload(domain.InboxItem{
		TicketID:      7,
		CustomerName:  "Maya",
		CustomerEmail: "maya@example.com",
		Subject:       "Where is my order?",
		Status:        "open",
		Priority:      "urgent",
		LastMessage:   "hello",
		LastMessageAt: lastMessageAt,
		MessageCount:  3,
	})

	if payload["message_count"] != 3 {
		t.Fatalf("message_count = %#v", payload["message_count"])
	}
	if payload["last_message_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("last_message_at = %#v", payload["last_message_at"])
	}
}
