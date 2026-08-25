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
	ID             int64      `json:"id"`
	OwnerID        int64      `json:"owner_id"`
	BusinessID     int64      `json:"business_id"`
	CustomerID     int64      `json:"customer_id"`
	Code           string     `json:"code"`
	InitialAmount  float64    `json:"initial_amount"`
	CurrentBalance float64    `json:"current_balance"`
	CustomerName   string     `json:"customer_name"`
	CustomerEmail  string     `json:"customer_email"`
	IsActive       bool       `json:"is_active"`
	ExpiresAt      *time.Time `json:"expires_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type GiftCardInput struct {
	PaidOrderID   int64   `json:"paid_order_id"`
	Code          string  `json:"code"`
	InitialAmount float64 `json:"initial_amount"`
	CustomerName  string  `json:"customer_name"`
	CustomerEmail string  `json:"customer_email"`
	ExpiresAt     string  `json:"expires_at"`
}

type RedeemInput struct {
	OrderID int64   `json:"order_id"`
	Code    string  `json:"code"`
	Amount  float64 `json:"amount"`
}

// CentsToFloat converts integer cents to standard float amount.
func CentsToFloat(cents int64) float64 {
	return float64(cents) / 100.0
}
