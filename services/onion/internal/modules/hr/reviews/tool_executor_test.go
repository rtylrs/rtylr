package reviews

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr/reviews/internal/domain"
)

func TestParseListReviewsToolInput(t *testing.T) {
	parsed, err := parseListReviewsToolInput(map[string]any{
		"status":   "draft",
		"period":   "2026-Q2",
		"staff_id": json.Number("4"),
		"page":     2,
		"limit":    "10",
	})
	if err != nil {
		t.Fatalf("parseListReviewsToolInput: %v", err)
	}
	if parsed.Status != "draft" || parsed.Period != "2026-Q2" || parsed.StaffID != 4 {
		t.Fatalf("parsed filters = %#v", parsed)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListReviewsToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListReviewsToolInput(map[string]any{"limit": 500})
	if err != nil {
		t.Fatalf("parseListReviewsToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListReviewsToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestFilterReviewsForTool(t *testing.T) {
	reviews := []domain.Review{
		{ID: 1, StaffID: 4, Period: "2026-Q2", Status: "draft"},
		{ID: 2, StaffID: 5, Period: "2026-Q2", Status: "draft"},
		{ID: 3, StaffID: 4, Period: "2026-Q1", Status: "submitted"},
	}
	filtered := filterReviewsForTool(reviews, listReviewsToolInput{
		Status:  "draft",
		Period:  "2026-Q2",
		StaffID: 4,
	})
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d", len(filtered))
	}
	if filtered[0].ID != 1 {
		t.Fatalf("filtered[0].ID = %d", filtered[0].ID)
	}
}

func TestReviewToolPayload(t *testing.T) {
	createdAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := reviewToolPayload(domain.Review{
		ID:        7,
		OwnerID:   2,
		StaffID:   4,
		Period:    "2026-Q2",
		Rating:    4.5,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	})

	if payload["rating"] != 4.5 {
		t.Fatalf("rating = %#v", payload["rating"])
	}
	if payload["created_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
}
