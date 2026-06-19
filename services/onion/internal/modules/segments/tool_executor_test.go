package segments

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/segments/internal/domain"
)

func TestParseListSegmentsToolInput(t *testing.T) {
	parsed, err := parseListSegmentsToolInput(map[string]any{
		"search":    "vip",
		"page":      json.Number("2"),
		"page_size": "10",
	})
	if err != nil {
		t.Fatalf("parseListSegmentsToolInput: %v", err)
	}
	if parsed.Query != "vip" {
		t.Fatalf("Query = %q", parsed.Query)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListSegmentsToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListSegmentsToolInput(map[string]any{"limit": 500})
	if err != nil {
		t.Fatalf("parseListSegmentsToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListSegmentsToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestFilterSegmentsForTool(t *testing.T) {
	segments := []domain.Segment{
		{ID: 1, Name: "VIP customers"},
		{ID: 2, Name: "Dormant"},
	}
	filtered := filterSegmentsForTool(segments, listSegmentsToolInput{Query: "vip"})
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d", len(filtered))
	}
	if filtered[0].ID != 1 {
		t.Fatalf("filtered[0].ID = %d", filtered[0].ID)
	}
}

func TestSegmentToolPayload(t *testing.T) {
	createdAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := segmentToolPayload(domain.Segment{
		ID:            7,
		OwnerID:       2,
		Name:          "VIP",
		Filters:       []byte(`{"tier":"gold"}`),
		CustomerCount: 12,
		CreatedAt:     createdAt,
		UpdatedAt:     createdAt,
	})

	if payload["customer_count"] != 12 {
		t.Fatalf("customer_count = %#v", payload["customer_count"])
	}
	if payload["created_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
}
