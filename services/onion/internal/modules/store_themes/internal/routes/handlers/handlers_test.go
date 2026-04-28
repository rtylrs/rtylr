package handlers

import (
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/store_themes/internal/domain"
)

func TestToPayload_RoundTrip(t *testing.T) {
	now := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	in := domain.Theme{
		ID:           5,
		OwnerID:      1001,
		Name:         "Modern Light",
		PrimaryColor: "#0ea5e9",
		Font:         "Inter",
		HeaderStyle:  "centered",
		FooterStyle:  "minimal",
		CustomCSS:    "body { background: white; }",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	out := toPayload(in)
	if out.ID != in.ID || out.Name != in.Name || out.PrimaryColor != in.PrimaryColor {
		t.Error("identity didn't round-trip")
	}
	if out.Font != in.Font || out.HeaderStyle != in.HeaderStyle || out.FooterStyle != in.FooterStyle {
		t.Error("style fields didn't round-trip")
	}
	if out.CustomCSS != in.CustomCSS {
		t.Error("custom_css didn't round-trip")
	}
	if !strings.HasPrefix(out.CreatedAt, "2026-04-28T") {
		t.Errorf("CreatedAt format unexpected: %q", out.CreatedAt)
	}
}
