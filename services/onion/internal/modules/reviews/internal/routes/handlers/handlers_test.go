package handlers

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/reviews/internal/domain"
)

func TestNewPayload_RoundTrip(t *testing.T) {
	now := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	in := domain.Review{
		ID:           4,
		OwnerID:      1001,
		StaffID:      4101,
		ReviewerID:   1002,
		Period:       "Q2-2026",
		Rating:       4.5,
		GoalsRating:  4,
		ValuesRating: 5,
		SkillsRating: 4,
		Notes:        "Great quarter",
		Status:       "submitted",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	out := newPayload(in)
	if out.ID != in.ID || out.StaffID != in.StaffID || out.ReviewerID != in.ReviewerID {
		t.Error("ids didn't round-trip")
	}
	if out.Rating != in.Rating || out.GoalsRating != in.GoalsRating || out.ValuesRating != in.ValuesRating || out.SkillsRating != in.SkillsRating {
		t.Errorf("ratings didn't round-trip: %+v", out)
	}
	if out.Period != in.Period || out.Notes != in.Notes || out.Status != in.Status {
		t.Error("period/notes/status didn't round-trip")
	}
}
