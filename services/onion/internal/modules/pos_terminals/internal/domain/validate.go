package domain

import (
	"fmt"
	"time"
)

// ValidateTerminalInput returns an error when the create input is malformed.
func ValidateTerminalInput(input TerminalInput) error {
	if input.BusinessID == 0 {
		return fmt.Errorf("business id is required")
	}
	if input.BranchID == 0 {
		return fmt.Errorf("branch id is required")
	}
	if input.Name == "" {
		return fmt.Errorf("terminal name is required")
	}
	return nil
}

// ValidateTerminalUpdateInput returns an error when the update input is malformed.
func ValidateTerminalUpdateInput(input TerminalUpdateInput) error {
	if input.ID == 0 {
		return fmt.Errorf("terminal id is required")
	}
	if input.BusinessID == 0 {
		return fmt.Errorf("business id is required")
	}
	if input.Name == "" {
		return fmt.Errorf("terminal name is required")
	}
	return nil
}

// IsOnline reports whether the terminal has recently been seen (within the last 5 minutes).
func (t Terminal) IsOnline() bool {
	if t.LastSeenAt == nil {
		return false
	}
	return time.Since(*t.LastSeenAt) <= 5*time.Minute
}
