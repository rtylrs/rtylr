package handlers

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/bills/internal/domain"
)

func TestNewPayload_AllFieldsRoundTrip(t *testing.T) {
	due := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	created := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	in := domain.Bill{
		ID:          7,
		OwnerID:     1001,
		SupplierID:  42,
		Amount:      999.99,
		Currency:    "USD",
		DueDate:     due,
		Status:      "open",
		Description: "Office rent",
		CreatedAt:   created,
		UpdatedAt:   created,
	}
	out := newPayload(in)

	if out.ID != in.ID || out.OwnerID != in.OwnerID || out.SupplierID != in.SupplierID {
		t.Errorf("identity didn't round-trip: %+v", out)
	}
	if out.Amount != in.Amount || out.Currency != in.Currency {
		t.Errorf("money didn't round-trip")
	}
	if !out.DueDate.Equal(in.DueDate) || !out.CreatedAt.Equal(in.CreatedAt) {
		t.Errorf("dates didn't round-trip")
	}
	if out.Status != in.Status || out.Description != in.Description {
		t.Errorf("status/description didn't round-trip")
	}
}
