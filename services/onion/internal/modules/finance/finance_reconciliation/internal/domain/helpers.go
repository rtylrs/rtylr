package domain

import "fmt"

// ValidateMatchInput checks structural requirements on a MatchInput before
// any database call is made.
func ValidateMatchInput(input MatchInput) error {
	if input.BankFeedID <= 0 {
		return fmt.Errorf("%w: bank_feed_id must be positive", ErrInvalidTarget)
	}

	if input.TargetID <= 0 {
		return fmt.Errorf("%w: target_id must be positive", ErrInvalidTarget)
	}

	if input.TargetType != TargetTypeInvoice && input.TargetType != TargetTypeExpense {
		return fmt.Errorf("%w: target_type must be 'invoice' or 'expense'", ErrInvalidTarget)
	}

	return nil
}

// ScoreDescription returns a human-readable label for a suggestion score
// bucket. It is useful for presenting the confidence of an auto-match to the
// operator UI.
func ScoreDescription(score float64) string {
	switch {
	case score == 0:
		return "exact"
	case score < 500:
		return "high"
	case score < 5000:
		return "medium"
	default:
		return "low"
	}
}
