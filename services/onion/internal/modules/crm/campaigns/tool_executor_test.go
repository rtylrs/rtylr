package campaigns

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/crm/campaigns/internal/domain"
)

func TestParseListCampaignsToolInput(t *testing.T) {
	parsed, err := parseListCampaignsToolInput(map[string]any{
		"search":     "summer",
		"status":     "sent",
		"type":       "email",
		"segment_id": json.Number("4"),
		"page":       2,
		"limit":      "10",
	})
	if err != nil {
		t.Fatalf("parseListCampaignsToolInput: %v", err)
	}
	if parsed.Query != "summer" || parsed.Status != "sent" || parsed.Type != "email" {
		t.Fatalf("parsed filters = %#v", parsed)
	}
	if parsed.SegmentID != 4 {
		t.Fatalf("SegmentID = %d", parsed.SegmentID)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListCampaignsToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListCampaignsToolInput(map[string]any{"limit": 500})
	if err != nil {
		t.Fatalf("parseListCampaignsToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListCampaignsToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestParseListCampaignsToolInputRejectsNegativeSegmentID(t *testing.T) {
	_, err := parseListCampaignsToolInput(map[string]any{"segment_id": -1})
	if err == nil || !strings.Contains(err.Error(), "segment_id") {
		t.Fatalf("error = %v, want segment_id validation", err)
	}
}

func TestFilterCampaignsForTool(t *testing.T) {
	segmentID := int64(4)
	campaigns := []domain.Campaign{
		{ID: 1, Name: "Summer Sale", Status: "sent", Type: "email", AudienceSegmentID: &segmentID},
		{ID: 2, Name: "Summer SMS", Status: "draft", Type: "sms", AudienceSegmentID: &segmentID},
		{ID: 3, Name: "Winter Sale", Status: "sent", Type: "email"},
	}

	filtered := filterCampaignsForTool(campaigns, listCampaignsToolInput{
		Query:     "summer",
		Status:    "sent",
		Type:      "email",
		SegmentID: 4,
	})
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d", len(filtered))
	}
	if filtered[0].ID != 1 {
		t.Fatalf("filtered[0].ID = %d", filtered[0].ID)
	}
}

func TestCampaignToolPayload(t *testing.T) {
	segmentID := int64(4)
	sentAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := campaignToolPayload(domain.Campaign{
		ID:                7,
		OwnerID:           2,
		Name:              "Summer",
		Type:              "email",
		Status:            "sent",
		AudienceSegmentID: &segmentID,
		SentCount:         20,
		OpenCount:         5,
		SentAt:            &sentAt,
		CreatedAt:         sentAt,
		UpdatedAt:         sentAt,
	})

	if payload["open_rate"] != 0.25 {
		t.Fatalf("open_rate = %#v", payload["open_rate"])
	}
	if payload["sent_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("sent_at = %#v", payload["sent_at"])
	}
}
