package domain

import "fmt"

// ValidateFloorInput returns an error when the create input is malformed.
func ValidateFloorInput(input FloorInput) error {
	if input.BusinessID == 0 {
		return fmt.Errorf("business id is required")
	}
	if input.BranchID == 0 {
		return fmt.Errorf("branch id is required")
	}
	if input.Name == "" {
		return fmt.Errorf("floor name is required")
	}
	if input.SortOrder < 0 {
		return fmt.Errorf("sort order cannot be negative")
	}
	return nil
}

// ValidateFloorUpdateInput returns an error when the update input is malformed.
func ValidateFloorUpdateInput(input FloorUpdateInput) error {
	if input.ID == 0 {
		return fmt.Errorf("floor id is required")
	}
	if input.BusinessID == 0 {
		return fmt.Errorf("business id is required")
	}
	if input.Name == "" {
		return fmt.Errorf("floor name is required")
	}
	return nil
}
