package domain

import (
	"database/sql"
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

func (r *RecurringInvoice) FromRow(row *sql.Rows) error {
	var endDate sql.NullTime

	err := row.Scan(
		&r.ID, &r.OwnerID, &r.TemplateInvoiceID, &r.Frequency,
		&r.NextInvoiceDate, &endDate, &r.IsActive, &r.CreatedAt,
	)
	if err != nil {
		return err
	}

	if endDate.Valid {
		r.EndDate = endDate.Time
	}

	return nil
}

type RecurringInvoiceInput struct {
	TemplateInvoiceID int64  `json:"template_invoice_id"`
	Frequency         string `json:"frequency"`
	NextInvoiceDate   string `json:"next_invoice_date"`
	EndDate           string `json:"end_date"`
	IsActive          bool   `json:"is_active"`
}
