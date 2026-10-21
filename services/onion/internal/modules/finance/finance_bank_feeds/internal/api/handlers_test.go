package api

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/finance/finance_bank_feeds/internal/domain"
)

func TestToPayload_RoundTripsAllFields(t *testing.T) {
	matchedExp := int64(7)
	matchedInv := int64(99)
	occurred := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	created := occurred.Add(-time.Hour)

	in := domain.BankFeedTransaction{
		ID:               1,
		OwnerID:          1001,
		BankAccountID:    50,
		AmountMinorUnits: -10050,
		Currency:         "USD",
		OccurredAt:       occurred,
		Description:      "Coffee",
		ExternalID:       "ext-abc",
		MatchedExpenseID: &matchedExp,
		MatchedInvoiceID: &matchedInv,
		CreatedAt:        created,
	}
	out := toPayload(in)
	if out.ID != in.ID || out.OwnerID != in.OwnerID || out.BankAccountID != in.BankAccountID {
		t.Error("ids didn't round-trip")
	}
	if out.AmountMinorUnits != in.AmountMinorUnits || out.Currency != in.Currency {
		t.Error("amount/currency didn't round-trip")
	}
	if !out.OccurredAt.Equal(in.OccurredAt) || !out.CreatedAt.Equal(in.CreatedAt) {
		t.Error("timestamps didn't round-trip")
	}
	if out.MatchedExpenseID == nil || *out.MatchedExpenseID != *in.MatchedExpenseID {
		t.Error("MatchedExpenseID didn't round-trip")
	}
	if out.MatchedInvoiceID == nil || *out.MatchedInvoiceID != *in.MatchedInvoiceID {
		t.Error("MatchedInvoiceID didn't round-trip")
	}
}

func TestToPayload_NilMatchPointersStayNil(t *testing.T) {
	out := toPayload(domain.BankFeedTransaction{})
	if out.MatchedExpenseID != nil || out.MatchedInvoiceID != nil {
		t.Errorf("nil pointers should stay nil: %+v", out)
	}
}
