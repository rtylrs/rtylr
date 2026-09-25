package domain

import (
	"errors"
	"net/mail"
	"strings"
)

// Validate returns an error if the SupplierInput is invalid.
func (i SupplierInput) Validate() error {
	if strings.TrimSpace(i.Name) == "" {
		return errors.New("supplier name is required")
	}
	if i.Email != "" {
		if _, err := mail.ParseAddress(i.Email); err != nil {
			return errors.New("supplier email is invalid")
		}
	}
	return nil
}
