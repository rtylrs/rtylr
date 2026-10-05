package supportcanned

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/support/canned/internal/domain"
)

func TestParseListSupportCannedResponsesToolInput(t *testing.T) {
	parsed, err := parseListSupportCannedResponsesToolInput(map[string]any{
		"search":    "refund",
		"category":  "billing",
		"page":      json.Number("2"),
		"page_size": "10",
	})
	if err != nil {
		t.Fatalf("parseListSupportCannedResponsesToolInput: %v", err)
	}
	if parsed.Query != "refund" {
		t.Fatalf("Query = %q", parsed.Query)
	}
	if parsed.Category != "billing" {
		t.Fatalf("Category = %q", parsed.Category)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListSupportCannedResponsesToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListSupportCannedResponsesToolInput(map[string]any{
		"limit": 500,
	})
	if err != nil {
		t.Fatalf("parseListSupportCannedResponsesToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListSupportCannedResponsesToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestParseListSupportCannedResponsesToolInputRejectsNegativeLimit(t *testing.T) {
	_, err := parseListSupportCannedResponsesToolInput(map[string]any{
		"limit": -1,
	})
	if err == nil || !strings.Contains(err.Error(), "limit") {
		t.Fatalf("error = %v, want limit validation", err)
	}
}

func TestFilterSupportCannedResponsesForTool(t *testing.T) {
	responses := []domain.CannedResponse{
		{ID: 1, Name: "Refund approved", Body: "We issued a refund", Category: "billing"},
		{ID: 2, Name: "Refund denied", Body: "We cannot refund", Category: "policy"},
		{ID: 3, Name: "Shipping update", Body: "Your order shipped", Category: "shipping"},
	}

	filtered := filterSupportCannedResponsesForTool(responses, listSupportCannedResponsesToolInput{
		Query:    "refund",
		Category: "billing",
	})
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d", len(filtered))
	}
	if filtered[0].ID != 1 {
		t.Fatalf("filtered[0].ID = %d", filtered[0].ID)
	}
}

func TestSupportCannedResponseToolPayload(t *testing.T) {
	createdAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := supportCannedResponseToolPayload(domain.CannedResponse{
		ID:         7,
		BusinessID: 2,
		Name:       "Refund approved",
		Category:   "billing",
		Shortcut:   "/refund-approved",
		CreatedAt:  createdAt,
	})

	if payload["shortcut"] != "/refund-approved" {
		t.Fatalf("shortcut = %#v", payload["shortcut"])
	}
	if payload["created_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
}
