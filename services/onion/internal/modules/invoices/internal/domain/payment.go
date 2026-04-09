package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrPaymentNotFound = errors.New("payment not found")

type Payment struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"ownerId"`
	InvoiceID   int64     `json:"invoiceId"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	PaymentDate time.Time `json:"paymentDate"`
	Method      string    `json:"method"`
	Reference   string    `json:"reference"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"createdAt"`
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
	InvoiceID   int64   `json:"invoiceId"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	PaymentDate string  `json:"paymentDate"`
	Method      string  `json:"method"`
	Reference   string  `json:"reference"`
	Notes       string  `json:"notes"`
}
