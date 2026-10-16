package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/platform/money"
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

	// Summed in minor units so the accounting identity is checked exactly.
	// This used to add float64 values and forgive a difference of up to half a
	// cent (const epsilon = 0.005), which meant an entry of 10.00 against
	// 9.996 was accepted as balanced and carried a third decimal the column
	// cannot hold (#1183). The guard was written in the same arithmetic it was
	// guarding.
	var totalDebit, totalCredit money.Amount
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
		if line.Debit > 0 && line.Credit > 0 {
			return fmt.Errorf("line[%d]: a journal line cannot have both debit and credit", i)
		}

		debit, err := money.FromFloat64(line.Debit)
		if err != nil {
			return fmt.Errorf("line[%d]: debit: %w", i, err)
		}
		credit, err := money.FromFloat64(line.Credit)
		if err != nil {
			return fmt.Errorf("line[%d]: credit: %w", i, err)
		}

		totalDebit += debit
		totalCredit += credit
	}

	if totalDebit != totalCredit {
		return fmt.Errorf("journal entry is not balanced: total_debit=%s total_credit=%s",
			totalDebit, totalCredit)
	}

	return nil
}
