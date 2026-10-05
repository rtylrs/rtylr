package domain

import (
	"testing"
	"time"
)

func TestFormatTotal_TwoDecimalPlaces(t *testing.T) {
	r := Receipt{Total: 42.5}
	if got := r.FormatTotal(); got != "42.50" {
		t.Errorf("FormatTotal: got %q, want '42.50'", got)
	}
}

func TestFormatTotal_ZeroTotal(t *testing.T) {
	r := Receipt{Total: 0}
	if got := r.FormatTotal(); got != "0.00" {
		t.Errorf("FormatTotal: got %q, want '0.00'", got)
	}
}

func TestFormatTotal_LargeAmount(t *testing.T) {
	r := Receipt{Total: 9999.99}
	if got := r.FormatTotal(); got != "9999.99" {
		t.Errorf("FormatTotal: got %q, want '9999.99'", got)
	}
}

func TestFormatReceiptHeader_ContainsOrderNumber(t *testing.T) {
	r := Receipt{OrderNumber: "ORD-001", PaidAt: time.Now()}
	header := r.FormatReceiptHeader()
	if header == "" {
		t.Fatal("expected non-empty header")
	}
	found := false
	for i := 0; i+len("ORD-001") <= len(header); i++ {
		if header[i:i+len("ORD-001")] == "ORD-001" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("header %q does not contain order number 'ORD-001'", header)
	}
}

func TestIsToday_SameDay(t *testing.T) {
	now := time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC)
	r := Receipt{PaidAt: now}
	if !r.IsToday(now) {
		t.Error("expected IsToday=true for same UTC day")
	}
}

func TestIsToday_DifferentDay(t *testing.T) {
	paidAt := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	now := time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC)
	r := Receipt{PaidAt: paidAt}
	if r.IsToday(now) {
		t.Error("expected IsToday=false for different UTC day")
	}
}

func TestIsToday_DifferentYear(t *testing.T) {
	paidAt := time.Date(2025, 5, 29, 12, 0, 0, 0, time.UTC)
	now := time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC)
	r := Receipt{PaidAt: paidAt}
	if r.IsToday(now) {
		t.Error("expected IsToday=false for different year")
	}
}
