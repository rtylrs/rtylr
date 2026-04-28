package handlers

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/journal/internal/domain"
)

func TestNewPayload_PostedEntryRoundTrip(t *testing.T) {
	now := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	in := domain.JournalEntry{
		ID:          7,
		OwnerID:     1001,
		Date:        now,
		Description: "Salary expense",
		Status:      "posted",
		PostedAt:    now.Add(time.Hour),
		Lines: []domain.JournalLine{
			{ID: 1, EntryID: 7, AccountID: 100, Debit: 1000, Credit: 0},
			{ID: 2, EntryID: 7, AccountID: 200, Debit: 0, Credit: 1000},
		},
		CreatedAt: now,
	}
	out := newPayload(in)

	if len(out.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(out.Lines))
	}
	if out.Lines[0].Debit != 1000 || out.Lines[1].Credit != 1000 {
		t.Error("debit/credit didn't round-trip")
	}
	if out.PostedAt == nil || !out.PostedAt.Equal(in.PostedAt) {
		t.Error("posted_at should be a non-nil pointer matching input")
	}
}

func TestNewPayload_DraftEntryHasNilPostedAt(t *testing.T) {
	in := domain.JournalEntry{
		Status: "draft",
		// PostedAt is the zero value.
	}
	out := newPayload(in)
	if out.PostedAt != nil {
		t.Errorf("expected nil PostedAt for unposted entry, got %v", *out.PostedAt)
	}
	if out.Lines == nil || len(out.Lines) != 0 {
		t.Errorf("expected empty lines slice (not nil), got %v", out.Lines)
	}
}
