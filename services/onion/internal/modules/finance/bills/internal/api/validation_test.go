package api

import (
	"testing"

	"github.com/voxire/onion/internal/modules/finance/bills/internal/domain"
)

func TestBillValidationRejectsNonPositiveAmount(t *testing.T) {
	if err := domain.ValidateBillInput(domain.BillInput{DueDate: "2026-01-01"}); err == nil {
		t.Fatal("expected non-positive amount to be rejected")
	}
}

func TestBillValidationRejectsInvalidStatus(t *testing.T) {
	if err := domain.ValidateBillInput(domain.BillInput{Amount: 100, DueDate: "2026-01-01", Status: "invalid"}); err == nil {
		t.Fatal("expected invalid status to be rejected")
	}
}
