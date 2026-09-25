package loyalty

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/loyalty/internal/domain"
)

func TestParseListLoyaltyTransactionsToolInput(t *testing.T) {
	parsed, err := parseListLoyaltyTransactionsToolInput(map[string]any{
		"customer_id": json.Number("4"),
		"page":        2,
		"limit":       "10",
	})
	if err != nil {
		t.Fatalf("parseListLoyaltyTransactionsToolInput: %v", err)
	}
	if parsed.CustomerID != 4 {
		t.Fatalf("CustomerID = %d", parsed.CustomerID)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListLoyaltyMembersToolInputCapsLimitAtServiceLimit(t *testing.T) {
	parsed, err := parseListLoyaltyMembersToolInput(map[string]any{"limit": 500})
	if err != nil {
		t.Fatalf("parseListLoyaltyMembersToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListLoyaltyMembersServiceLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestParseListLoyaltyRewardsToolInput(t *testing.T) {
	parsed, err := parseListLoyaltyRewardsToolInput(map[string]any{"active_only": "active"})
	if err != nil {
		t.Fatalf("parseListLoyaltyRewardsToolInput: %v", err)
	}
	if !parsed.ActiveOnly {
		t.Fatal("ActiveOnly = false")
	}
}

func TestLoyaltyConfigToolPayload(t *testing.T) {
	updatedAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := loyaltyConfigToolPayload(domain.LoyaltyConfig{
		ID:              7,
		OwnerID:         2,
		PointsPerDollar: 1.5,
		IsActive:        true,
		UpdatedAt:       updatedAt,
	})

	if payload["points_per_dollar"] != 1.5 {
		t.Fatalf("points_per_dollar = %#v", payload["points_per_dollar"])
	}
	if payload["updated_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("updated_at = %#v", payload["updated_at"])
	}
}

func TestLoyaltyRewardToolPayload(t *testing.T) {
	stock := 3
	payload := loyaltyRewardToolPayload(domain.LoyaltyReward{
		ID:         7,
		OwnerID:    2,
		Name:       "Free Shake",
		PointsCost: 100,
		Stock:      &stock,
		Active:     true,
	})

	if payload["stock"] != 3 {
		t.Fatalf("stock = %#v", payload["stock"])
	}
}

func TestLoyaltyTransactionToolPayload(t *testing.T) {
	referenceID := int64(9)
	createdAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := loyaltyTransactionToolPayload(domain.LoyaltyTransaction{
		ID:          7,
		OwnerID:     2,
		CustomerID:  4,
		Points:      50,
		ReferenceID: &referenceID,
		CreatedAt:   createdAt,
	})

	if payload["reference_id"] != int64(9) {
		t.Fatalf("reference_id = %#v", payload["reference_id"])
	}
	if payload["created_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
}
