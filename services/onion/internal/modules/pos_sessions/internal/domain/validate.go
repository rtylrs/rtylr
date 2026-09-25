package domain

import "fmt"

// ValidateOpenSessionInput returns an error when the input is malformed.
func ValidateOpenSessionInput(input OpenSessionInput) error {
	if input.OwnerID == 0 {
		return fmt.Errorf("owner id is required")
	}
	if input.OpeningFloat < 0 {
		return fmt.Errorf("opening float cannot be negative")
	}
	return nil
}

// ValidateCloseSessionInput returns an error when the input is malformed.
func ValidateCloseSessionInput(input CloseSessionInput) error {
	if input.ID == 0 {
		return fmt.Errorf("session id is required")
	}
	if input.OwnerID == 0 {
		return fmt.Errorf("owner id is required")
	}
	if input.ClosingFloat < 0 {
		return fmt.Errorf("closing float cannot be negative")
	}
	return nil
}

// IsOpen reports whether the session has open status.
func (s Session) IsOpen() bool {
	return s.Status == "open"
}

// IsClosed reports whether the session has closed status.
func (s Session) IsClosed() bool {
	return s.Status == "closed"
}
