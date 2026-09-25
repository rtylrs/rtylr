package api

import (
	"context"
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
		BusinessID:  2001,
		Name:        "Test Shop",
		Domain:      "shop.example.com",
		LogoURL:     "https://example.com/logo.png",
		ThemeID:     5,
		IsPublished: true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	out := toPayload(in)
	if out.ID != in.ID || out.OwnerID != in.OwnerID || out.BusinessID != in.BusinessID || out.ThemeID != in.ThemeID {
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

func TestToPayload_UnpublishedStorefront(t *testing.T) {
	now := time.Now()
	in := domain.Storefront{ID: 2, Name: "Draft", IsPublished: false, CreatedAt: now, UpdatedAt: now}
	out := toPayload(in)
	if out.IsPublished {
		t.Error("expected IsPublished=false")
	}
	if out.Name != "Draft" {
		t.Errorf("name not preserved: %q", out.Name)
	}
}

func TestCreateHandler_Validate_RequiresName(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when name is empty")
	}
}

func TestCreateHandler_Validate_WhitespaceName(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "   "}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for whitespace name")
	}
}

func TestCreateHandler_Validate_HappyPath(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "My Store"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateHandler_Validate_WithDomain(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "My Store", Domain: "store.example.com"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateHandler_Validate_RequiresID(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 0, Name: "Updated"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when ID is zero")
	}
}

func TestUpdateHandler_Validate_RequiresName(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 3, Name: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when name is empty")
	}
}

func TestUpdateHandler_Validate_HappyPath(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 5, Name: "Updated Store"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateHandler_Validate_WhitespaceName(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 1, Name: "  \t  "}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for whitespace name")
	}
}

func TestToPayload_ZeroThemeID(t *testing.T) {
	now := time.Now()
	in := domain.Storefront{ID: 3, Name: "No Theme", ThemeID: 0, CreatedAt: now, UpdatedAt: now}
	out := toPayload(in)
	if out.ThemeID != 0 {
		t.Errorf("expected ThemeID=0, got %d", out.ThemeID)
	}
}

func TestToPayload_TimestampFormat(t *testing.T) {
	now := time.Date(2026, 1, 15, 8, 30, 0, 0, time.UTC)
	in := domain.Storefront{ID: 4, Name: "TS", CreatedAt: now, UpdatedAt: now}
	out := toPayload(in)
	if !strings.HasPrefix(out.CreatedAt, "2026-01-15T08:30:00") {
		t.Errorf("unexpected timestamp format: %q", out.CreatedAt)
	}
}
