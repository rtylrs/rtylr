package domain

import "fmt"

// ValidateLocationInput returns an error when the create input is malformed.
func ValidateLocationInput(input LocationInput) error {
	if input.OwnerID == 0 {
		return fmt.Errorf("owner id is required")
	}
	if input.Name == "" {
		return fmt.Errorf("location name is required")
	}
	return nil
}

// ValidateLocationUpdateInput returns an error when the update input is malformed.
func ValidateLocationUpdateInput(input LocationUpdateInput) error {
	if input.ID == 0 {
		return fmt.Errorf("location id is required")
	}
	if input.OwnerID == 0 {
		return fmt.Errorf("owner id is required")
	}
	if input.Name == "" {
		return fmt.Errorf("location name is required")
	}
	return nil
}
