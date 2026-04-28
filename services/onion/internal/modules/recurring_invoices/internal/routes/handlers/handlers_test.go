package handlers

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/recurring_invoices/internal/domain"
)

func TestNewPayload_RoundTrip(t *testing.T) {
	now := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	in := domain.RecurringInvoice{
		ID:                3,
		OwnerID:           1001,
		TemplateInvoiceID: 99,
		Frequency:         "monthly",
		NextInvoiceDate:   now.AddDate(0, 1, 0),
		EndDate:           now.AddDate(1, 0, 0),
		IsActive:          true,
		CreatedAt:         now,
	}
	out := newPayload(in)
	if out.ID != in.ID || out.TemplateInvoiceID != in.TemplateInvoiceID {
		t.Error("ids didn't round-trip")
	}
	if out.Frequency != in.Frequency || !out.NextInvoiceDate.Equal(in.NextInvoiceDate) {
		t.Error("frequency / next date didn't round-trip")
	}
	if !out.EndDate.Equal(in.EndDate) || !out.IsActive {
		t.Error("end date / is_active didn't round-trip")
	}
}
