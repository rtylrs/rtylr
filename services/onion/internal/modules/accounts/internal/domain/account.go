package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrAccountNotFound = errors.New("account not found")

type Account struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"ownerId"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Currency  string    `json:"currency"`
	Balance   float64   `json:"balance"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
}

func (a *Account) FromRow(row *sql.Rows) error {
	var (
		currency sql.NullString
	)

	err := row.Scan(
		&a.ID, &a.OwnerID, &a.Name, &a.Type, &currency,
		&a.Balance, &a.IsActive, &a.CreatedAt,
	)
	if err != nil {
		return err
	}

	if currency.Valid {
		a.Currency = currency.String
	}

	return nil
}

type AccountInput struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
	IsActive bool    `json:"isActive"`
}
