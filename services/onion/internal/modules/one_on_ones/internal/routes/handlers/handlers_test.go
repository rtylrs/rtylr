package handlers

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/one_on_ones/internal/domain"
)

func TestNewPayload_RoundTrip(t *testing.T) {
	now := time.Date(2026, 4, 28, 10, 0, 0, 0, time.UTC)
	in := domain.OneOnOne{
		ID:          1,
		OwnerID:     1001,
		ManagerID:   1002,
		StaffID:     4101,
		ScheduledAt: now,
		Notes:       "Discuss Q2 goals",
		ActionItems: "Sign up for training",
		Status:      "scheduled",
		CreatedAt:   now.Add(-time.Hour),
	}
	out := newPayload(in)
	if out.ID != in.ID || out.ManagerID != in.ManagerID || out.StaffID != in.StaffID {
		t.Error("ids didn't round-trip")
	}
	if !out.ScheduledAt.Equal(in.ScheduledAt) {
		t.Errorf("scheduled_at didn't round-trip: %v vs %v", out.ScheduledAt, in.ScheduledAt)
	}
	if out.Notes != in.Notes || out.ActionItems != in.ActionItems || out.Status != in.Status {
		t.Error("text fields didn't round-trip")
	}
}
