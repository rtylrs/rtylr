package domain

import (
	"errors"
	"time"
)

var ErrSupplierNotFound = errors.New("supplier not found")

type Supplier struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"owner_id"`
	BusinessID   int64     `json:"business_id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Address      string    `json:"address"`
	City         string    `json:"city"`
	Country      string    `json:"country"`
	TaxID        string    `json:"tax_id"`
	PaymentTerms string    `json:"payment_terms"`
	Notes        string    `json:"notes"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
type SupplierInput struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Address      string `json:"address"`
	City         string `json:"city"`
	Country      string `json:"country"`
	TaxID        string `json:"tax_id"`
	PaymentTerms string `json:"payment_terms"`
	Notes        string `json:"notes"`
	IsActive     bool   `json:"is_active"`
}
