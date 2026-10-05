package supportknowledge

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/support/knowledge/internal/domain"
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

func TestParseGetSupportArticleToolInputAcceptsAliases(t *testing.T) {
	cases := []map[string]any{
		{"article_id": int64(7)},
		{"support_article_id": json.Number("7")},
		{"knowledge_article_id": float64(7)},
		{"article": "#7"},
		{"id": "support-article-7"},
		{"id": "knowledge_article_7"},
		{"id": "kb-7"},
		{"slug": " Refund-Policy "},
		{"title": "Refund Policy"},
	}

	for _, input := range cases {
		parsed, err := parseGetSupportArticleToolInput(input)
		if err != nil {
			t.Fatalf("parseGetSupportArticleToolInput(%v) returned error: %v", input, err)
		}
		if parsed.ArticleID == 0 && parsed.Slug == "" && parsed.Title == "" {
			t.Fatalf("parsed empty input for %v: %+v", input, parsed)
		}
	}
}

func TestParseGetSupportArticleToolInputNormalizesSlug(t *testing.T) {
	parsed, err := parseGetSupportArticleToolInput(map[string]any{"slug": " Refund-Policy "})
	if err != nil {
		t.Fatalf("parseGetSupportArticleToolInput: %v", err)
	}
	if parsed.Slug != "refund-policy" {
		t.Fatalf("Slug = %q", parsed.Slug)
	}
}

func TestParseGetSupportArticleToolInputRejectsMissingLookup(t *testing.T) {
	_, err := parseGetSupportArticleToolInput(map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "article_id") {
		t.Fatalf("error = %v, want lookup validation", err)
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
