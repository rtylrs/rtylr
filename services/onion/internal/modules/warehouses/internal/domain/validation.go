package domain

import "errors"

// Validate returns an error if the WarehouseInput is invalid.
func (i WarehouseInput) Validate() error {
	if i.Name == "" {
		return errors.New("warehouse name is required")
	}
	if i.Capacity < 0 {
		return errors.New("warehouse capacity must not be negative")
	}
	return nil
}
