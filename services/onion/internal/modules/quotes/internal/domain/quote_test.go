package domain

import (
	"testing"
	"time"
)

func TestIsExpired_TrueWhenPastValidUntilAndNotTerminal(t *testing.T) {
	q := &Quote{
		Status:     "sent",
		ValidUntil: time.Now().AddDate(0, 0, -1),
	}
	if !q.IsExpired() {
		t.Error("expected IsExpired=true for past ValidUntil on sent quote")
	}
}

func TestIsExpired_FalseWhenAccepted(t *testing.T) {
	q := &Quote{
		Status:     "accepted",
		ValidUntil: time.Now().AddDate(0, 0, -5),
	}
	if q.IsExpired() {
		t.Error("expected IsExpired=false for accepted quote even if past ValidUntil")
	}
}

func TestIsExpired_FalseWhenDeclined(t *testing.T) {
	q := &Quote{
		Status:     "declined",
		ValidUntil: time.Now().AddDate(0, 0, -5),
	}
	if q.IsExpired() {
		t.Error("expected IsExpired=false for declined quote")
	}
}

func TestIsExpired_FalseWhenFutureValidUntil(t *testing.T) {
	q := &Quote{
		Status:     "sent",
		ValidUntil: time.Now().AddDate(0, 0, 10),
	}
	if q.IsExpired() {
		t.Error("expected IsExpired=false when ValidUntil is in the future")
	}
}

func TestIsExpired_FalseWhenZeroValidUntil(t *testing.T) {
	q := &Quote{Status: "draft"}
	if q.IsExpired() {
		t.Error("expected IsExpired=false when ValidUntil is zero")
	}
}

func TestIsEditable_TrueForDraft(t *testing.T) {
	q := &Quote{Status: "draft"}
	if !q.IsEditable() {
		t.Error("expected IsEditable=true for draft")
	}
}

func TestIsEditable_TrueForSent(t *testing.T) {
	q := &Quote{Status: "sent"}
	if !q.IsEditable() {
		t.Error("expected IsEditable=true for sent")
	}
}

func TestIsEditable_FalseForAccepted(t *testing.T) {
	q := &Quote{Status: "accepted"}
	if q.IsEditable() {
		t.Error("expected IsEditable=false for accepted")
	}
}

func TestIsEditable_FalseForDeclined(t *testing.T) {
	q := &Quote{Status: "declined"}
	if q.IsEditable() {
		t.Error("expected IsEditable=false for declined")
	}
}

func TestNetTotal_SubtotalMinusDiscountPlusTax(t *testing.T) {
	q := &Quote{Subtotal: 1000, Discount: 100, Tax: 50}
	want := 950.0
	if got := q.NetTotal(); got != want {
		t.Errorf("expected NetTotal=%f, got %f", want, got)
	}
}

func TestNetTotal_ZeroWhenAllZero(t *testing.T) {
	q := &Quote{}
	if got := q.NetTotal(); got != 0 {
		t.Errorf("expected 0, got %f", got)
	}
}
