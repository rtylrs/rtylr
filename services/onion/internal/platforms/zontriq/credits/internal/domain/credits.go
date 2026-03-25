package domain

import (
	"errors"
	"time"

	"github.com/voxire/onion/internal/platform/utils/convert"
	sqlc "github.com/voxire/onion/internal/platforms/zontriq/credits/internal/infrastructure/persistence/sqlc"
)

var (
	ErrCreditsNotFound     = errors.New("credits not found")
	ErrInsufficientCredits = errors.New("insufficient credits")
	ErrUserNotFound        = errors.New("zontriq user not found")
)

type Credits struct {
	ID             int32
	ZontriqUserID  int32
	Balance        int32
	TotalPurchased int32
	TotalSpent     int32
	Status         int32
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (c *Credits) FromRows(row sqlc.ZontriqCredit) error {
	c.ID = row.ID
	c.ZontriqUserID = row.ZontriqUserID
	c.Balance = row.Balance
	c.TotalPurchased = row.TotalPurchased
	c.TotalSpent = row.TotalSpent
	c.Status = row.Status
	c.CreatedAt = row.CreatedAt.Time
	c.UpdatedAt = row.UpdatedAt.Time

	return nil
}

type CreditTransaction struct {
	ID            int32
	ZontriqUserID int32
	Amount        int32
	Type          string
	Description   string
	ReferenceType string
	ReferenceID   int32
	Status        int32
	CreatedAt     time.Time
}

func (t *CreditTransaction) FromRows(row sqlc.ZontriqCreditTransaction) error {
	t.ID = row.ID
	t.ZontriqUserID = row.ZontriqUserID
	t.Amount = row.Amount
	t.Type = row.Type
	t.Description = convert.ToString(row.Description)
	t.ReferenceType = convert.ToString(row.ReferenceType)
	t.ReferenceID = convert.ToInt32(row.ReferenceID)
	t.Status = row.Status
	t.CreatedAt = row.CreatedAt.Time

	return nil
}
