package infrastructure

import (
	"database/sql"
	"testing"
)

func TestPaymentAmountToCents(t *testing.T) {
	tests := []struct {
		name   string
		amount float64
		want   int64
	}{
		{name: "issue scenario first payment", amount: 18.33, want: 1833},
		{name: "whole dollars", amount: 36, want: 3600},
		{name: "single cent", amount: 0.01, want: 1},
		{name: "zero", amount: 0, want: 0},
		{name: "negative", amount: -5.25, want: -525},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := int64(paymentAmountToCents(tt.amount)); got != tt.want {
				t.Errorf("paymentAmountToCents(%v) = %d, want %d", tt.amount, got, tt.want)
			}
		})
	}
}

func TestNullDecimalToCents(t *testing.T) {
	tests := []struct {
		name  string
		value sql.NullString
		want  int64
	}{
		{name: "decimal string", value: sql.NullString{String: "36.66", Valid: true}, want: 3666},
		{name: "zero", value: sql.NullString{String: "0.00", Valid: true}, want: 0},
		{name: "null column", value: sql.NullString{}, want: 0},
		{name: "empty string", value: sql.NullString{String: "", Valid: true}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := int64(nullDecimalToCents(tt.value)); got != tt.want {
				t.Errorf("nullDecimalToCents(%v) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}

// TestPaidStatusArithmetic reproduces issue #757's exact figures: a $36.66
// invoice paid with two $18.33 payments must land on exactly zero due.
// In float64 the same accumulation leaves ~+2.4e-15 and the invoice stays
// partial forever.
func TestPaidStatusArithmetic(t *testing.T) {
	totalCents := nullDecimalToCents(sql.NullString{String: "36.66", Valid: true})
	paidCents := nullDecimalToCents(sql.NullString{String: "0.00", Valid: true})

	paidCents += paymentAmountToCents(18.33)
	if newDue := totalCents - paidCents; newDue <= 0 {
		t.Fatalf("invoice marked paid after one of two payments: due=%d", newDue)
	}

	paidCents += paymentAmountToCents(18.33)
	newDue := totalCents - paidCents
	if newDue > 0 {
		t.Errorf("fully paid invoice still has due=%d cents; status would stay partial", newDue)
	}
	if newDue < 0 {
		t.Errorf("due went negative without clamping: %d", newDue)
	}
}
