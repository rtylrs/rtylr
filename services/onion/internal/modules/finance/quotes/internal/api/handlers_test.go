package api

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/finance/quotes/internal/domain"
)

func TestNewPayload_RoundTrip(t *testing.T) {
	now := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	in := domain.Quote{
		ID:            5,
		OwnerID:       1001,
		CustomerName:  "Acme",
		CustomerEmail: "buyer@acme.test",
		Status:        "sent",
		Items:         `[{"sku":"x","qty":1}]`,
		Subtotal:      100,
		Tax:           10,
		Discount:      5,
		Total:         105,
		ValidUntil:    now.AddDate(0, 1, 0),
		Notes:         "30-day payment terms",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	out := newPayload(in)

	if out.ID != in.ID || out.OwnerID != in.OwnerID {
		t.Error("ids didn't round-trip")
	}
	if out.CustomerName != in.CustomerName || out.CustomerEmail != in.CustomerEmail || out.Status != in.Status {
		t.Error("customer/status didn't round-trip")
	}
	if out.Subtotal != in.Subtotal || out.Tax != in.Tax || out.Discount != in.Discount || out.Total != in.Total {
		t.Error("money didn't round-trip")
	}
	if out.Items != in.Items || out.Notes != in.Notes {
		t.Error("items/notes didn't round-trip")
	}
	if !out.ValidUntil.Equal(in.ValidUntil) || !out.CreatedAt.Equal(in.CreatedAt) {
		t.Error("dates didn't round-trip")
	}
}
