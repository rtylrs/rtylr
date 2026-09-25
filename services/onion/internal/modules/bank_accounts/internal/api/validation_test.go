package api

import (
	"testing"

	"github.com/voxire/onion/internal/modules/bank_accounts/internal/domain"
)

func TestBankAccountValidationRejectsMissingName(t *testing.T) {
	if err := domain.ValidateBankAccountInput(domain.BankAccountInput{}); err == nil {
		t.Fatal("expected missing name to be rejected")
	}
}
