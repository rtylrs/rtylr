package domain

import "strings"

// HasContactInfo reports whether the supplier has at least one form of contact information.
func (s Supplier) HasContactInfo() bool {
	return s.Email != "" || s.Phone != ""
}

// HasAddress reports whether the supplier has a full address.
func (s Supplier) HasAddress() bool {
	return strings.TrimSpace(s.Address) != "" && strings.TrimSpace(s.City) != ""
}

// DisplayName returns the supplier name trimmed of whitespace.
func (s Supplier) DisplayName() string {
	return strings.TrimSpace(s.Name)
}
