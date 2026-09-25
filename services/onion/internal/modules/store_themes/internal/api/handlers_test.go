package api

import (
	"context"
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

func TestToPayload_EmptyCustomCSS(t *testing.T) {
	now := time.Now()
	in := domain.Theme{ID: 1, Name: "Minimal", CustomCSS: "", CreatedAt: now, UpdatedAt: now}
	out := toPayload(in)
	if out.CustomCSS != "" {
		t.Errorf("expected empty CustomCSS, got %q", out.CustomCSS)
	}
}

func TestToPayload_TimestampFormat(t *testing.T) {
	now := time.Date(2026, 2, 10, 14, 0, 0, 0, time.UTC)
	in := domain.Theme{ID: 2, Name: "T", CreatedAt: now, UpdatedAt: now}
	out := toPayload(in)
	if !strings.HasPrefix(out.CreatedAt, "2026-02-10T14:00:00") {
		t.Errorf("unexpected timestamp: %q", out.CreatedAt)
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
	req := &CreateRequest{Name: "Dark Mode", PrimaryColor: "#000"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateHandler_Validate_WithOptionalFields(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{
		Name:         "Ocean",
		PrimaryColor: "#0ea5e9",
		Font:         "Roboto",
		HeaderStyle:  "wide",
		FooterStyle:  "full",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateHandler_Validate_RequiresID(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 0}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when ID is zero")
	}
}

func TestUpdateHandler_Validate_HappyPath(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 3}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateHandler_Validate_LargeID(t *testing.T) {
	h := &UpdateHandler{}
	req := &UpdateRequest{ID: 99999}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error for large id: %v", err)
	}
}

func TestToPayload_OwnerIDPreserved(t *testing.T) {
	now := time.Now()
	in := domain.Theme{ID: 7, OwnerID: 42, Name: "X", CreatedAt: now, UpdatedAt: now}
	out := toPayload(in)
	if out.OwnerID != 42 {
		t.Errorf("expected OwnerID=42, got %d", out.OwnerID)
	}
}
