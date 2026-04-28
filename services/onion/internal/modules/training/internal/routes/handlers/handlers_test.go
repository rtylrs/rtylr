package handlers

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/training/internal/domain"
)

func TestNewPayload_RoundTrip(t *testing.T) {
	now := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	in := domain.TrainingCourse{
		ID:            12,
		OwnerID:       1001,
		Name:          "Onboarding 101",
		Description:   "Welcome to the team",
		DurationHours: 4.5,
		IsRequired:    true,
		CreatedAt:     now,
	}
	out := newPayload(in)
	if out.ID != in.ID || out.Name != in.Name || out.Description != in.Description {
		t.Error("identity didn't round-trip")
	}
	if out.DurationHours != in.DurationHours {
		t.Errorf("duration didn't round-trip: %v vs %v", out.DurationHours, in.DurationHours)
	}
	if out.IsRequired != in.IsRequired || !out.CreatedAt.Equal(in.CreatedAt) {
		t.Error("required/created_at didn't round-trip")
	}
}
