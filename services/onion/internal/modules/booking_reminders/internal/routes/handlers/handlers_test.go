package handlers

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/booking_reminders/internal/domain"
)

func TestToPayload_FormatsTimestampsAsRFC3339(t *testing.T) {
	scheduled := time.Date(2026, 4, 28, 14, 30, 0, 0, time.UTC)
	sent := time.Date(2026, 4, 28, 14, 35, 0, 0, time.UTC)
	created := time.Date(2026, 4, 27, 10, 0, 0, 0, time.UTC)

	in := domain.Reminder{
		ID:            5,
		OwnerID:       1001,
		AppointmentID: 88,
		Type:          "email",
		ScheduledAt:   scheduled,
		SentAt:        sent,
		CreatedAt:     created,
	}
	out := toPayload(in)

	if out.ID != in.ID || out.OwnerID != in.OwnerID || out.AppointmentID != in.AppointmentID {
		t.Error("ids didn't round-trip")
	}
	if out.Type != in.Type {
		t.Errorf("type didn't round-trip: %q", out.Type)
	}
	if !strings.HasPrefix(out.ScheduledAt, "2026-04-28T14:30:00") {
		t.Errorf("scheduled_at format unexpected: %q", out.ScheduledAt)
	}
}

func TestCreateHandler_Validate_RequiresAppointmentID(t *testing.T) {
	h := &CreateHandler{}
	if err := h.Validate(context.Background(), &CreateRequest{Type: "email"}); err == nil {
		t.Error("expected validation error for missing appointment id")
	} else if err.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", err.Code)
	}
}

func TestCreateHandler_Validate_AcceptsValidRequest(t *testing.T) {
	h := &CreateHandler{}
	if err := h.Validate(context.Background(), &CreateRequest{AppointmentID: 1, Type: "sms"}); err != nil {
		t.Errorf("unexpected error: %+v", err)
	}
}
