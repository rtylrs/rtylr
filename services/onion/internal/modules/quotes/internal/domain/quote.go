package domain

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrQuoteNotFound         = errors.New("quote not found")
	ErrQuoteAlreadyConverted = errors.New("quote already converted")
)

type Quote struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"owner_id"`
	CustomerName  string    `json:"customer_name"`
	CustomerEmail string    `json:"customer_email"`
	Status        string    `json:"status"`
	Items         string    `json:"items"`
	Subtotal      float64   `json:"subtotal"`
	Tax           float64   `json:"tax"`
	Discount      float64   `json:"discount"`
	Total         float64   `json:"total"`
	ValidUntil    time.Time `json:"valid_until"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (q *Quote) FromRow(row *sql.Rows) error {
	var (
		customerEmail sql.NullString
		items         sql.NullString
		validUntil    sql.NullTime
		notes         sql.NullString
	)

	err := row.Scan(
		&q.ID, &q.OwnerID, &q.CustomerName, &customerEmail, &q.Status,
		&items, &q.Subtotal, &q.Tax, &q.Discount, &q.Total,
		&validUntil, &notes, &q.CreatedAt, &q.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if customerEmail.Valid {
		q.CustomerEmail = customerEmail.String
	}
	if items.Valid {
		q.Items = items.String
	}
	if validUntil.Valid {
		q.ValidUntil = validUntil.Time
	}
	if notes.Valid {
		q.Notes = notes.String
	}

	return nil
}

type QuoteInput struct {
	CustomerName  string  `json:"customer_name"`
	CustomerEmail string  `json:"customer_email"`
	Status        string  `json:"status"`
	Items         string  `json:"items"`
	Subtotal      float64 `json:"subtotal"`
	Tax           float64 `json:"tax"`
	Discount      float64 `json:"discount"`
	Total         float64 `json:"total"`
	ValidUntil    string  `json:"valid_until"`
	Notes         string  `json:"notes"`
}
