package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrBankAccountNotFound = errors.New("bank account not found")

type BankAccount struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"ownerId"`
	Name         string    `json:"name"`
	Institution  string    `json:"institution"`
	AccountType  string    `json:"accountType"`
	Balance      float64   `json:"balance"`
	Currency     string    `json:"currency"`
	IsConnected  bool      `json:"isConnected"`
	LastSyncedAt time.Time `json:"lastSyncedAt"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (b *BankAccount) FromRow(row *sql.Rows) error {
	var (
		institution  sql.NullString
		accountType  sql.NullString
		currency     sql.NullString
		lastSyncedAt sql.NullTime
	)

	err := row.Scan(
		&b.ID, &b.OwnerID, &b.Name, &institution, &accountType,
		&b.Balance, &currency, &b.IsConnected, &lastSyncedAt, &b.CreatedAt,
	)
	if err != nil {
		return err
	}

	if institution.Valid {
		b.Institution = institution.String
	}
	if accountType.Valid {
		b.AccountType = accountType.String
	}
	if currency.Valid {
		b.Currency = currency.String
	}
	if lastSyncedAt.Valid {
		b.LastSyncedAt = lastSyncedAt.Time
	}

	return nil
}

type BankAccountInput struct {
	Name        string  `json:"name"`
	Institution string  `json:"institution"`
	AccountType string  `json:"accountType"`
	Balance     float64 `json:"balance"`
	Currency    string  `json:"currency"`
	IsConnected bool    `json:"isConnected"`
}
