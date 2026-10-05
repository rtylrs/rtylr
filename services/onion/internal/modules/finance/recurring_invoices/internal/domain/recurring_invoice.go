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
	ID                int64
	OwnerID           int64
	BusinessID        int64
	TemplateInvoiceID int64
	Frequency         string
	NextInvoiceDate   time.Time
	EndDate           time.Time
	IsActive          bool
	CreatedAt         time.Time
}

// RecurringInvoiceInput carries the writable recurring invoice fields. The
// pointer fields keep the contract's optional fields distinguishable from a
// zero value: nil means the caller did not send the field, so create applies
// its default and update leaves the stored value unchanged.
type RecurringInvoiceInput struct {
	TemplateInvoiceID *int64
	Frequency         string
	NextInvoiceDate   string
	EndDate           *string
	IsActive          *bool
}
