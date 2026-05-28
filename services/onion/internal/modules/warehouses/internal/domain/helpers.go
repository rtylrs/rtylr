package domain

import "strings"

// HasLocation reports whether the warehouse has city or country info.
func (w Warehouse) HasLocation() bool {
	return strings.TrimSpace(w.City) != "" || strings.TrimSpace(w.Country) != ""
}

// HasManager reports whether a manager name is set on the warehouse.
func (w Warehouse) HasManager() bool {
	return strings.TrimSpace(w.ManagerName) != ""
}

// HasCapacity reports whether the warehouse has a positive capacity set.
func (w Warehouse) HasCapacity() bool {
	return w.Capacity > 0
}

// DisplayName returns the warehouse name trimmed of whitespace.
func (w Warehouse) DisplayName() string {
	return strings.TrimSpace(w.Name)
}
