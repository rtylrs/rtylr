package supportknowledge

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/support_knowledge/internal/domain"
)

func TestParseListSupportArticlesToolInput(t *testing.T) {
	parsed, err := parseListSupportArticlesToolInput(map[string]any{
		"page":      json.Number("2"),
		"page_size": "10",
	})
	if err != nil {
		t.Fatalf("parseListSupportArticlesToolInput: %v", err)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListSupportArticlesToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListSupportArticlesToolInput(map[string]any{
		"limit": 500,
	})
	if err != nil {
		t.Fatalf("parseListSupportArticlesToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListSupportArticlesToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestSupportArticleToolPayload(t *testing.T) {
	createdAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := supportArticleToolPayload(domain.Article{
		ID:         7,
		BusinessID: 2,
		Title:      "Refund policy",
		Slug:       "refund-policy",
		Category:   "billing",
		Views:      12,
		CreatedAt:  createdAt,
		UpdatedAt:  createdAt,
	})

	if payload["views"] != 12 {
		t.Fatalf("views = %#v", payload["views"])
	}
	if payload["created_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
}
