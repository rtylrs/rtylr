package domain

import (
	"fmt"
	"strings"
	"time"
)

// ValidateJournalEntryInput returns a non-nil error if the input fails
// structural validation before touching the database.
func ValidateJournalEntryInput(input JournalEntryInput) error {
	if strings.TrimSpace(input.Date) == "" {
		return fmt.Errorf("date is required")
	}

	if _, err := time.Parse("2006-01-02", input.Date); err != nil {
		return fmt.Errorf("date must be in YYYY-MM-DD format")
	}

	if len(input.Lines) < 2 {
		return fmt.Errorf("a journal entry must have at least two lines")
	}

	var totalDebit, totalCredit float64
	for i, line := range input.Lines {
		if line.AccountID <= 0 {
			return fmt.Errorf("line[%d]: account_id must be positive", i)
		}
		if line.Debit < 0 {
			return fmt.Errorf("line[%d]: debit must not be negative", i)
		}
		if line.Credit < 0 {
			return fmt.Errorf("line[%d]: credit must not be negative", i)
		}
		if line.Debit == 0 && line.Credit == 0 {
			return fmt.Errorf("line[%d]: at least one of debit or credit must be non-zero", i)
		}
		totalDebit += line.Debit
		totalCredit += line.Credit
	}

	// Accounting identity: debits must equal credits (within floating-point epsilon).
	const epsilon = 0.005
	diff := totalDebit - totalCredit
	if diff < -epsilon || diff > epsilon {
		return fmt.Errorf("journal entry is not balanced: total_debit=%.2f total_credit=%.2f", totalDebit, totalCredit)
	}

	return nil
}

// CentsToFloat converts integer cents to standard float amount.
func CentsToFloat(cents int64) float64 {
	return float64(cents) / 100.0
}
