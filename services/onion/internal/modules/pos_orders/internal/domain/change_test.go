package domain

import (
	"errors"
	"testing"
)

func TestComputeChange_ExactPayment(t *testing.T) {
	result, err := ComputeChange(50.00, 50.00)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(result.Change, 0) {
		t.Errorf("change: got %g, want 0", result.Change)
	}
}

func TestComputeChange_Overpayment(t *testing.T) {
	result, err := ComputeChange(30.00, 50.00)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(result.Change, 20.00) {
		t.Errorf("change: got %g, want 20", result.Change)
	}
}

func TestComputeChange_Underpayment_Errors(t *testing.T) {
	_, err := ComputeChange(100.00, 80.00)
	if err == nil {
		t.Fatal("expected error for underpayment")
	}
	if !errors.Is(err, ErrPaymentInsufficient) {
		t.Fatalf("expected ErrPaymentInsufficient, got %v", err)
	}
}

func TestComputeChange_ZeroTotal_ZeroChange(t *testing.T) {
	result, err := ComputeChange(0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Change != 0 {
		t.Errorf("expected zero change, got %g", result.Change)
	}
}

func TestComputeChange_AmountDueField(t *testing.T) {
	result, err := ComputeChange(75.50, 100.00)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(result.AmountDue, 75.50) {
		t.Errorf("AmountDue: got %g, want 75.50", result.AmountDue)
	}
}

func TestComputeChange_AmountPaidField(t *testing.T) {
	result, err := ComputeChange(20.00, 50.00)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(result.AmountPaid, 50.00) {
		t.Errorf("AmountPaid: got %g, want 50", result.AmountPaid)
	}
}
