package domain

import (
	"errors"
	"time"
)

var (
	ErrGiftCardNotFound      = errors.New("gift card not found")
	ErrGiftCardInactive      = errors.New("gift card is inactive")
	ErrGiftCardExpired       = errors.New("gift card has expired")
	ErrInsufficientBalance   = errors.New("insufficient gift card balance")
)

type GiftCard struct {
	ID             int64      `json:"id"`
	OwnerID        int64      `json:"ownerId"`
	Code           string     `json:"code"`
	InitialAmount  float64    `json:"initialAmount"`
	CurrentBalance float64    `json:"currentBalance"`
	CustomerName   string     `json:"customerName"`
	CustomerEmail  string     `json:"customerEmail"`
	IsActive       bool       `json:"isActive"`
	ExpiresAt      *time.Time `json:"expiresAt"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type GiftCardInput struct {
	Code           string  `json:"code"`
	InitialAmount  float64 `json:"initialAmount"`
	CustomerName   string  `json:"customerName"`
	CustomerEmail  string  `json:"customerEmail"`
	ExpiresAt      string  `json:"expiresAt"`
}

type RedeemInput struct {
	Code   string  `json:"code"`
	Amount float64 `json:"amount"`
}
