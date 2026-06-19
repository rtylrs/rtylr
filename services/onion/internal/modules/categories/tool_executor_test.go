package categories

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/categories/internal/domain"
)

func TestParseCreateCategoryToolInput(t *testing.T) {
	parsed, err := parseCreateCategoryToolInput(map[string]any{
		"category_name": " Supplements ",
		"description":   " Sports nutrition ",
	})
	if err != nil {
		t.Fatalf("parseCreateCategoryToolInput: %v", err)
	}
	if parsed.Name != "Supplements" {
		t.Fatalf("Name = %q", parsed.Name)
	}
	if parsed.Description != "Sports nutrition" {
		t.Fatalf("Description = %q", parsed.Description)
	}
}

func TestParseCreateCategoryToolInputRejectsMissingName(t *testing.T) {
	_, err := parseCreateCategoryToolInput(map[string]any{"description": "No name"})
	if err == nil {
		t.Fatal("expected missing name error")
	}
}

func TestCategoryToolPayloadFormatsFields(t *testing.T) {
	createdAt := time.Date(2026, 6, 18, 8, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := categoryToolPayload(domain.Category{
		ID:          9,
		Name:        "Supplements",
		Description: "Sports nutrition",
		CreatedAt:   createdAt,
	})

	if payload["id"] != int64(9) {
		t.Fatalf("id = %v", payload["id"])
	}
	if payload["name"] != "Supplements" {
		t.Fatalf("name = %v", payload["name"])
	}
	if payload["description"] != "Sports nutrition" {
		t.Fatalf("description = %v", payload["description"])
	}
	if payload["created_at"] != "2026-06-18T05:00:00Z" {
		t.Fatalf("created_at = %v", payload["created_at"])
	}
}
