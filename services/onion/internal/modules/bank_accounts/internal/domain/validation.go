package domain

import (
	"fmt"
	"strings"
)

// ValidAccountTypes lists account types accepted by the bank_accounts module.
var ValidAccountTypes = map[string]struct{}{
	"checking": {},
	"savings":  {},
	"credit":   {},
	"loan":     {},
	"other":    {},
}

// ValidateBankAccountInput returns a non-nil error when the input violates
// a rule that can be checked without touching the database.
func ValidateBankAccountInput(input BankAccountInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("name is required")
	}

	if input.Balance < 0 {
		return fmt.Errorf("balance must not be negative")
	}

	if input.AccountType != "" {
		if _, ok := ValidAccountTypes[input.AccountType]; !ok {
			return fmt.Errorf("invalid account_type %q", input.AccountType)
		}
	}

	return nil
}
