package domain

import (
	"fmt"
	"strings"
)

// ValidateSearchQuery sanitises a customer search query.
// Empty queries are allowed (returns all). Queries over 200 chars are rejected.
func ValidateSearchQuery(query string) error {
	if len(strings.TrimSpace(query)) > 200 {
		return fmt.Errorf("search query exceeds maximum length of 200 characters")
	}
	return nil
}

// DisplayName returns a single display name from first and last name.
func (c Customer) DisplayName() string {
	if c.LastName == "" {
		return c.FirstName
	}
	return c.FirstName + " " + c.LastName
}

// HasContactInfo returns true when the customer has at least one contact field set.
func (c Customer) HasContactInfo() bool {
	return c.Email != "" || c.Phone != ""
}
