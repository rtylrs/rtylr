package domain

import (
	"errors"
	"time"
)

var (
	ErrGiftCardNotFound    = errors.New("gift card not found")
	ErrGiftCardInactive    = errors.New("gift card is inactive")
	ErrGiftCardExpired     = errors.New("gift card has expired")
	ErrInsufficientBalance = errors.New("insufficient gift card balance")
	ErrInvalidTender       = errors.New("gift card tender reference is invalid")
	ErrInvalidOrder        = errors.New("gift card order reference is invalid")
)

type GiftCard struct {
	ID             int64
	OwnerID        int64
	BusinessID     int64
	CustomerID     int64
	Code           string
	InitialAmount  float64
	CurrentBalance float64
	CustomerName   string
	CustomerEmail  string
	IsActive       bool
	ExpiresAt      *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type GiftCardInput struct {
	PaidOrderID   int64
	Code          string
	InitialAmount float64
	CustomerName  string
	CustomerEmail string
	ExpiresAt     string
}

type RedeemInput struct {
	OrderID int64
	Code    string
	Amount  float64
}

// MaskCode keeps only the last four characters of a gift card code. The code
// is the bearer credential for redemption, so it must not reach logs in full.
func MaskCode(code string) string {
	runes := []rune(code)
	if len(runes) <= 4 {
		return "****"
	}
	return "****" + string(runes[len(runes)-4:])
}
