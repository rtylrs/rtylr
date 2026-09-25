package domain

import (
	"errors"
	"time"
)

var (
	ErrRecurringInvoiceNotFound        = errors.New("recurring invoice not found")
	ErrRecurringInvoiceTemplateMissing = errors.New("recurring invoice template not found")
)

type RecurringInvoice struct {
	ID                int64     `json:"id"`
	OwnerID           int64     `json:"owner_id"`
	BusinessID        int64     `json:"business_id"`
	TemplateInvoiceID int64     `json:"template_invoice_id"`
	Frequency         string    `json:"frequency"`
	NextInvoiceDate   time.Time `json:"next_invoice_date"`
	EndDate           time.Time `json:"end_date"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
}

// RecurringInvoiceInput carries the writable recurring invoice fields. The
// pointer fields keep the contract's optional fields distinguishable from a
// zero value: nil means the caller did not send the field, so create applies
// its default and update leaves the stored value unchanged.
type RecurringInvoiceInput struct {
	TemplateInvoiceID *int64  `json:"template_invoice_id,omitempty"`
	Frequency         string  `json:"frequency"`
	NextInvoiceDate   string  `json:"next_invoice_date"`
	EndDate           *string `json:"end_date,omitempty"`
	IsActive          *bool   `json:"is_active,omitempty"`
}
