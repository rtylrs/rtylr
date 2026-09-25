package domain

import "fmt"

// ValidTableStatuses lists the valid statuses for a table.
var ValidTableStatuses = []string{"available", "occupied", "reserved", "cleaning"}

// ValidateTableInput returns an error when the create input is malformed.
func ValidateTableInput(input TableInput) error {
	if input.OwnerID == 0 {
		return fmt.Errorf("owner id is required")
	}
	if input.Label == "" {
		return fmt.Errorf("table label is required")
	}
	if input.Seats < 0 {
		return fmt.Errorf("seats cannot be negative")
	}
	return nil
}

// ValidateTableUpdateInput returns an error when the update input is malformed.
func ValidateTableUpdateInput(input TableUpdateInput) error {
	if input.ID == 0 {
		return fmt.Errorf("table id is required")
	}
	if input.OwnerID == 0 {
		return fmt.Errorf("owner id is required")
	}
	if input.Label == "" {
		return fmt.Errorf("table label is required")
	}
	if input.Status != "" {
		for _, s := range ValidTableStatuses {
			if input.Status == s {
				return nil
			}
		}
		return fmt.Errorf("invalid status %q", input.Status)
	}
	return nil
}

func ValidateTableStatusInput(input TableStatusInput) error {
	if input.ID == 0 {
		return fmt.Errorf("table id is required")
	}
	if input.OwnerID == 0 {
		return fmt.Errorf("owner id is required")
	}
	for _, status := range ValidTableStatuses {
		if input.Status == status {
			return nil
		}
	}
	return fmt.Errorf("invalid status %q", input.Status)
}

// IsAvailable reports whether the table can accept new customers.
func (t Table) IsAvailable() bool {
	return t.Status == "available"
}

// IsOccupied reports whether the table currently has customers.
func (t Table) IsOccupied() bool {
	return t.Status == "occupied"
}
