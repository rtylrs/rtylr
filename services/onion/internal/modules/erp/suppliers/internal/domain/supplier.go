package domain

import (
	"errors"
	"time"
)

var ErrSupplierNotFound = errors.New("supplier not found")

type Supplier struct {
	ID           int64
	OwnerID      int64
	BusinessID   int64
	Name         string
	Email        string
	Phone        string
	Address      string
	City         string
	Country      string
	TaxID        string
	PaymentTerms string
	Notes        string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
type SupplierInput struct {
	Name         string
	Email        string
	Phone        string
	Address      string
	City         string
	Country      string
	TaxID        string
	PaymentTerms string
	Notes        string
	IsActive     bool
}
