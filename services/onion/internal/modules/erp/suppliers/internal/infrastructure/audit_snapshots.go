package infrastructure

import (
	"time"

	"github.com/voxire/onion/internal/modules/erp/suppliers/internal/domain"
)

type supplierSnapshot struct {
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

type supplierInputSnapshot struct {
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

func supplierSnapshotOf(value domain.Supplier) supplierSnapshot { return supplierSnapshot(value) }

func supplierInputSnapshotOf(value domain.SupplierInput) supplierInputSnapshot {
	return supplierInputSnapshot(value)
}
