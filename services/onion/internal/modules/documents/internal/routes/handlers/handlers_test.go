package handlers

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/documents/internal/domain"
)

func TestNewPayload_RoundTrip(t *testing.T) {
	now := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	in := domain.Document{
		ID:         11,
		OwnerID:    1001,
		StaffID:    4101,
		Name:       "ID Scan.pdf",
		Type:       "id",
		FileURL:    "https://r2.example/test/scan.pdf",
		ExpiryDate: now.AddDate(2, 0, 0),
		CreatedAt:  now,
	}
	out := newPayload(in)
	if out.ID != in.ID || out.OwnerID != in.OwnerID || out.StaffID != in.StaffID {
		t.Errorf("ids didn't round-trip")
	}
	if out.Name != in.Name || out.Type != in.Type || out.FileURL != in.FileURL {
		t.Errorf("metadata didn't round-trip")
	}
	if !out.ExpiryDate.Equal(in.ExpiryDate) || !out.CreatedAt.Equal(in.CreatedAt) {
		t.Errorf("dates didn't round-trip")
	}
}
