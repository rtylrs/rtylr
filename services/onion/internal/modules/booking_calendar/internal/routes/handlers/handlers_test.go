package handlers

import (
	"context"
	"testing"
)

func TestCalendarHandler_Validate_RequiresFrom(t *testing.T) {
	h := &CalendarHandler{}
	req := &CalendarRequest{From: "", To: "2026-06-30"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for missing from")
	}
}

func TestCalendarHandler_Validate_RequiresTo(t *testing.T) {
	h := &CalendarHandler{}
	req := &CalendarRequest{From: "2026-06-01", To: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for missing to")
	}
}

func TestCalendarHandler_Validate_RejectsWhitespaceFrom(t *testing.T) {
	h := &CalendarHandler{}
	req := &CalendarRequest{From: "   ", To: "2026-06-30"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for whitespace-only from")
	}
}

func TestCalendarHandler_Validate_RejectsWhitespaceTo(t *testing.T) {
	h := &CalendarHandler{}
	req := &CalendarRequest{From: "2026-06-01", To: "   "}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for whitespace-only to")
	}
}

func TestCalendarHandler_Validate_AcceptsValidRequest(t *testing.T) {
	h := &CalendarHandler{}
	req := &CalendarRequest{From: "2026-06-01", To: "2026-06-30"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCalendarHandler_Validate_TrimsFromAndTo(t *testing.T) {
	h := &CalendarHandler{}
	req := &CalendarRequest{From: "  2026-06-01  ", To: "  2026-06-30  "}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.From != "2026-06-01" {
		t.Errorf("expected trimmed From, got %q", req.From)
	}
	if req.To != "2026-06-30" {
		t.Errorf("expected trimmed To, got %q", req.To)
	}
}
