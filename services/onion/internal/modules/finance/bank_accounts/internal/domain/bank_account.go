package domain

import (
	"errors"
	"time"

	"github.com/voxire/onion/internal/platform/money"
)

var ErrBankAccountNotFound = errors.New("bank account not found")

type BankAccount struct {
	ID           int64
	OwnerID      int64
	BusinessID   int64
	Name         string
	Institution  string
	AccountType  string
	LastFour     string
	Balance      money.Amount
	Currency     string
	IsPrimary    bool
	IsConnected  bool
	LastSyncedAt time.Time
	CreatedAt    time.Time
}
type BankAccountInput struct {
	Name        string
	Institution string
	AccountType string
	LastFour    string
	Balance     money.Amount
	Currency    string
	IsPrimary   bool
	IsConnected bool
}
