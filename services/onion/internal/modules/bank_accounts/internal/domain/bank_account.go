package domain

import (
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/platform/money"
)

var ErrBankAccountNotFound = errors.New("bank account not found")

type BankAccount struct {
	ID           int64        `json:"id"`
	OwnerID      int64        `json:"owner_id"`
	BusinessID   int64        `json:"business_id"`
	Name         string       `json:"name"`
	Institution  string       `json:"institution"`
	AccountType  string       `json:"account_type"`
	Balance      money.Amount `json:"balance"`
	Currency     string       `json:"currency"`
	IsConnected  bool         `json:"is_connected"`
	LastSyncedAt time.Time    `json:"last_synced_at"`
	CreatedAt    time.Time    `json:"created_at"`
}

func (b *BankAccount) FromRow(row *sql.Rows) error {
	var (
		institution  sql.NullString
		accountType  sql.NullString
		currency     sql.NullString
		lastSyncedAt sql.NullTime
	)

	err := row.Scan(
		&b.ID, &b.OwnerID, &b.BusinessID, &b.Name, &institution, &accountType,
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
	Name        string       `json:"name"`
	Institution string       `json:"institution"`
	AccountType string       `json:"account_type"`
	Balance     money.Amount `json:"balance"`
	Currency    string       `json:"currency"`
	IsConnected bool         `json:"is_connected"`
}
