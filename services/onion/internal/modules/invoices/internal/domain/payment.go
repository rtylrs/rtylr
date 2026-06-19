package domain

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrInvalidPaymentAmount = errors.New("payment amount must be greater than zero")
)

type Payment struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	BusinessID  int64     `json:"business_id"`
	InvoiceID   int64     `json:"invoice_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	PaymentDate time.Time `json:"payment_date"`
	Method      string    `json:"method"`
	Reference   string    `json:"reference"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
}

func (p *Payment) FromRow(row *sql.Rows) error {
	var (
		currency  sql.NullString
		reference sql.NullString
		notes     sql.NullString
	)

	err := row.Scan(
		&p.ID,
		&p.OwnerID,
		&p.BusinessID,
		&p.InvoiceID,
		&p.Amount,
		&currency,
		&p.PaymentDate,
		&p.Method,
		&reference,
		&notes,
		&p.CreatedAt,
	)
	if err != nil {
		return err
	}

	if currency.Valid {
		p.Currency = currency.String
	} else {
		p.Currency = "USD"
	}

	if reference.Valid {
		p.Reference = reference.String
	}

	if notes.Valid {
		p.Notes = notes.String
	}

	return nil
}

type PaymentInput struct {
	InvoiceID   int64   `json:"invoice_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	PaymentDate string  `json:"payment_date"`
	Method      string  `json:"method"`
	Reference   string  `json:"reference"`
	Notes       string  `json:"notes"`
}
