package domain

import (
	"testing"
)

func ptr64(v int64) *int64 { return &v }

func TestIsMatched_TrueWhenExpenseIDSet(t *testing.T) {
	txn := &BankFeedTransaction{MatchedExpenseID: ptr64(1)}
	if !txn.IsMatched() {
		t.Error("expected IsMatched=true when MatchedExpenseID is set")
	}
}

func TestIsMatched_TrueWhenInvoiceIDSet(t *testing.T) {
	txn := &BankFeedTransaction{MatchedInvoiceID: ptr64(5)}
	if !txn.IsMatched() {
		t.Error("expected IsMatched=true when MatchedInvoiceID is set")
	}
}

func TestIsMatched_FalseWhenNeitherSet(t *testing.T) {
	txn := &BankFeedTransaction{}
	if txn.IsMatched() {
		t.Error("expected IsMatched=false when neither ID is set")
	}
}

func TestAmountDecimal_ConvertsMinorUnits(t *testing.T) {
	txn := &BankFeedTransaction{AmountMinorUnits: 12345}
	want := 123.45
	if got := txn.AmountDecimal(100); got != want {
		t.Errorf("expected %f, got %f", want, got)
	}
}

func TestAmountDecimal_DefaultDivisor100WhenZero(t *testing.T) {
	txn := &BankFeedTransaction{AmountMinorUnits: 500}
	want := 5.0
	if got := txn.AmountDecimal(0); got != want {
		t.Errorf("expected %f with zero divisor, got %f", want, got)
	}
}

func TestIsCredit_TrueForPositiveAmount(t *testing.T) {
	txn := &BankFeedTransaction{AmountMinorUnits: 1000}
	if !txn.IsCredit() {
		t.Error("expected IsCredit=true for positive amount")
	}
}

func TestIsCredit_FalseForNegativeAmount(t *testing.T) {
	txn := &BankFeedTransaction{AmountMinorUnits: -500}
	if txn.IsCredit() {
		t.Error("expected IsCredit=false for negative amount")
	}
}

func TestIsDebit_TrueForNegativeAmount(t *testing.T) {
	txn := &BankFeedTransaction{AmountMinorUnits: -200}
	if !txn.IsDebit() {
		t.Error("expected IsDebit=true for negative amount")
	}
}

func TestIsDebit_FalseForPositiveAmount(t *testing.T) {
	txn := &BankFeedTransaction{AmountMinorUnits: 300}
	if txn.IsDebit() {
		t.Error("expected IsDebit=false for positive amount")
	}
}

func TestIsCredit_FalseForZero(t *testing.T) {
	txn := &BankFeedTransaction{AmountMinorUnits: 0}
	if txn.IsCredit() {
		t.Error("expected IsCredit=false for zero amount")
	}
}

func TestIsDebit_FalseForZero(t *testing.T) {
	txn := &BankFeedTransaction{AmountMinorUnits: 0}
	if txn.IsDebit() {
		t.Error("expected IsDebit=false for zero amount")
	}
}
