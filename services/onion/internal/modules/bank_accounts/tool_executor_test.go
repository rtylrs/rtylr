package bankaccounts

import (
	"strings"
	"testing"
)

func TestParseGetBankAccountToolInputAcceptsAliases(t *testing.T) {
	cases := []map[string]any{
		{"bank_account_id": int64(12)},
		{"account_id": float64(12)},
		{"account": "#12"},
		{"id": "bank-account-12"},
		{"id": "account_12"},
		{"id": "acct-12"},
	}

	for _, input := range cases {
		parsed, err := parseGetBankAccountToolInput(input)
		if err != nil {
			t.Fatalf("parseGetBankAccountToolInput(%v) returned error: %v", input, err)
		}
		if parsed.BankAccountID != 12 {
			t.Fatalf("BankAccountID = %d, want 12 for %v", parsed.BankAccountID, input)
		}
	}
}

func TestParseGetBankAccountToolInputRejectsMissingAccount(t *testing.T) {
	_, err := parseGetBankAccountToolInput(map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "bank_account_id") {
		t.Fatalf("error = %v, want bank_account_id validation", err)
	}
}
