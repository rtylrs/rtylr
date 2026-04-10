package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrRecurringInvoiceNotFound = errors.New("recurring invoice not found")

type RecurringInvoice struct {
	ID                int64     `json:"id"`
	OwnerID           int64     `json:"ownerId"`
	TemplateInvoiceID int64     `json:"templateInvoiceId"`
	Frequency         string    `json:"frequency"`
	NextInvoiceDate   time.Time `json:"nextInvoiceDate"`
	EndDate           time.Time `json:"endDate"`
	IsActive          bool      `json:"isActive"`
	CreatedAt         time.Time `json:"createdAt"`
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
	TemplateInvoiceID int64  `json:"templateInvoiceId"`
	Frequency         string `json:"frequency"`
	NextInvoiceDate   string `json:"nextInvoiceDate"`
	EndDate           string `json:"endDate"`
	IsActive          bool   `json:"isActive"`
}
