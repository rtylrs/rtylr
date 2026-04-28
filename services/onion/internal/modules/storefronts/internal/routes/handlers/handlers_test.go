package handlers

import (
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/storefronts/internal/domain"
)

func TestToPayload_RoundTrip(t *testing.T) {
	now := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	in := domain.Storefront{
		ID:          1,
		OwnerID:     1001,
		Name:        "Test Shop",
		Domain:      "shop.example.com",
		LogoURL:     "https://example.com/logo.png",
		ThemeID:     5,
		IsPublished: true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	out := toPayload(in)
	if out.ID != in.ID || out.OwnerID != in.OwnerID || out.ThemeID != in.ThemeID {
		t.Error("ids didn't round-trip")
	}
	if out.Name != in.Name || out.Domain != in.Domain || out.LogoURL != in.LogoURL {
		t.Error("strings didn't round-trip")
	}
	if !strings.HasPrefix(out.CreatedAt, "2026-04-28T12:00:00") {
		t.Errorf("CreatedAt format unexpected: %q", out.CreatedAt)
	}
	if !out.IsPublished {
		t.Error("is_published didn't round-trip")
	}
}
