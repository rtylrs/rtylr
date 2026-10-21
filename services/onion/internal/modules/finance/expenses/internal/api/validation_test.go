package api

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/finance/expenses/internal/domain"
)

func TestExpenseValidationRejectsMissingDescription(t *testing.T) {
	if err := domain.ValidatExpenseInput(1, "", "", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("expected missing description to be rejected")
	}
}

func TestExpenseValidationRejectsMissingDate(t *testing.T) {
	if err := domain.ValidatExpenseInput(1, "", "valid", time.Time{}); err == nil {
		t.Fatal("expected missing date to be rejected")
	}
}
