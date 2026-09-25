package domain

import (
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
	LastFour     string       `json:"last_four"`
	Balance      money.Amount `json:"balance"`
	Currency     string       `json:"currency"`
	IsPrimary    bool         `json:"is_primary"`
	IsConnected  bool         `json:"is_connected"`
	LastSyncedAt time.Time    `json:"last_synced_at"`
	CreatedAt    time.Time    `json:"created_at"`
}
type BankAccountInput struct {
	Name        string       `json:"name"`
	Institution string       `json:"institution"`
	AccountType string       `json:"account_type"`
	LastFour    string       `json:"last_four"`
	Balance     money.Amount `json:"balance"`
	Currency    string       `json:"currency"`
	IsPrimary   bool         `json:"is_primary"`
	IsConnected bool         `json:"is_connected"`
}
