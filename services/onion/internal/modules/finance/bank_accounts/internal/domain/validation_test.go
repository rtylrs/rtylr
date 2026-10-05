package domain

import (
	"testing"

	"github.com/voxire/onion/internal/platform/money"
)

func TestValidateBankAccountInput_HappyPath(t *testing.T) {
	input := BankAccountInput{Name: "Main Checking", AccountType: "checking", Balance: money.Amount(100000)}
	if err := ValidateBankAccountInput(input); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateBankAccountInput_MissingName(t *testing.T) {
	input := BankAccountInput{AccountType: "checking"}
	if err := ValidateBankAccountInput(input); err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestValidateBankAccountInput_NegativeBalance(t *testing.T) {
	input := BankAccountInput{Name: "Savings", Balance: money.Amount(-100)}
	if err := ValidateBankAccountInput(input); err == nil {
		t.Fatal("expected error for negative balance")
	}
}

func TestValidateBankAccountInput_InvalidAccountType(t *testing.T) {
	input := BankAccountInput{Name: "Test", AccountType: "unknown"}
	if err := ValidateBankAccountInput(input); err == nil {
		t.Fatal("expected error for invalid account_type")
	}
}

func TestValidateBankAccountInput_EmptyAccountTypeOK(t *testing.T) {
	input := BankAccountInput{Name: "Test"}
	if err := ValidateBankAccountInput(input); err != nil {
		t.Errorf("unexpected error for empty account_type: %v", err)
	}
}

func TestValidateBankAccountInput_ValidAccountTypes(t *testing.T) {
	for atype := range ValidAccountTypes {
		input := BankAccountInput{Name: "Account", AccountType: atype}
		if err := ValidateBankAccountInput(input); err != nil {
			t.Errorf("unexpected error for account_type=%q: %v", atype, err)
		}
	}
}

func TestValidateBankAccountInput_ZeroBalanceOK(t *testing.T) {
	input := BankAccountInput{Name: "New Account", Balance: money.Amount(0)}
	if err := ValidateBankAccountInput(input); err != nil {
		t.Errorf("unexpected error for zero balance: %v", err)
	}
}
