package brands

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/brands/internal/domain"
)

func TestParseCreateBrandToolInput(t *testing.T) {
	parsed, err := parseCreateBrandToolInput(map[string]any{
		"brand_name":  " Optimum Nutrition ",
		"logo":        " https://example.com/on.png ",
		"description": " Protein supplements ",
	})
	if err != nil {
		t.Fatalf("parseCreateBrandToolInput: %v", err)
	}
	if parsed.Name != "Optimum Nutrition" {
		t.Fatalf("Name = %q", parsed.Name)
	}
	if parsed.LogoURL != "https://example.com/on.png" {
		t.Fatalf("LogoURL = %q", parsed.LogoURL)
	}
	if parsed.Description != "Protein supplements" {
		t.Fatalf("Description = %q", parsed.Description)
	}
}

func TestParseCreateBrandToolInputRejectsMissingName(t *testing.T) {
	_, err := parseCreateBrandToolInput(map[string]any{"description": "No name"})
	if err == nil {
		t.Fatal("expected missing name error")
	}
}

func TestBrandToolPayloadFormatsFields(t *testing.T) {
	createdAt := time.Date(2026, 6, 18, 9, 15, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := brandToolPayload(domain.Brand{
		ID:           4,
		Name:         "Optimum Nutrition",
		LogoURL:      "https://example.com/on.png",
		Description:  "Protein supplements",
		ProductCount: 12,
		CreatedAt:    createdAt,
	})

	if payload["id"] != int64(4) {
		t.Fatalf("id = %v", payload["id"])
	}
	if payload["name"] != "Optimum Nutrition" {
		t.Fatalf("name = %v", payload["name"])
	}
	if payload["product_count"] != int64(12) {
		t.Fatalf("product_count = %v", payload["product_count"])
	}
	if payload["created_at"] != "2026-06-18T06:15:00Z" {
		t.Fatalf("created_at = %v", payload["created_at"])
	}
}
