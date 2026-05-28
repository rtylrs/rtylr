package domain

import "testing"

func TestIsTerminal_PaidIsTerminal(t *testing.T) {
	if !IsTerminal(StatusPaid) {
		t.Error("expected paid to be terminal")
	}
}

func TestIsTerminal_VoidIsTerminal(t *testing.T) {
	if !IsTerminal(StatusVoid) {
		t.Error("expected void to be terminal")
	}
}

func TestIsTerminal_DraftIsNotTerminal(t *testing.T) {
	if IsTerminal(StatusDraft) {
		t.Error("draft should not be terminal")
	}
}

func TestIsTerminal_SentIsNotTerminal(t *testing.T) {
	if IsTerminal(StatusSent) {
		t.Error("sent should not be terminal")
	}
}

func TestIsValidStatus_AllValidStatuses(t *testing.T) {
	for s := range ValidInvoiceStatuses {
		if !IsValidStatus(s) {
			t.Errorf("expected IsValidStatus(%q)=true", s)
		}
	}
}

func TestIsValidStatus_UnknownStatusFalse(t *testing.T) {
	for _, s := range []string{"unknown", "deleted", "", "PAID"} {
		if IsValidStatus(s) {
			t.Errorf("expected IsValidStatus(%q)=false", s)
		}
	}
}

func TestStatusConstants_AreNonEmpty(t *testing.T) {
	for _, s := range []string{
		StatusDraft, StatusSent, StatusViewed,
		StatusPartial, StatusPaid, StatusOverdue, StatusVoid,
	} {
		if s == "" {
			t.Error("status constant should not be empty")
		}
	}
}

func TestPaymentMethodConstants_AreNonEmpty(t *testing.T) {
	for _, m := range []string{
		PaymentMethodCash, PaymentMethodCard,
		PaymentMethodBankTransfer, PaymentMethodOther,
	} {
		if m == "" {
			t.Error("payment method constant should not be empty")
		}
	}
}
