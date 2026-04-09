package domain

import (
	"errors"
	"time"
)

var (
	ErrLoyaltyConfigNotFound  = errors.New("loyalty config not found")
	ErrInsufficientPoints     = errors.New("insufficient loyalty points")
	ErrTransactionNotFound    = errors.New("loyalty transaction not found")
)

type LoyaltyTransaction struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"ownerId"`
	CustomerID    int64     `json:"customerId"`
	Points        int       `json:"points"`
	Type          string    `json:"type"`
	Description   string    `json:"description"`
	ReferenceID   *int64    `json:"referenceId"`
	ReferenceType string    `json:"referenceType"`
	CreatedAt     time.Time `json:"createdAt"`
}

type LoyaltyConfig struct {
	ID                   int64     `json:"id"`
	OwnerID              int64     `json:"ownerId"`
	PointsPerDollar      float64   `json:"pointsPerDollar"`
	MinRedemptionPoints  int       `json:"minRedemptionPoints"`
	PointsValue          float64   `json:"pointsValue"`
	IsActive             bool      `json:"isActive"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

type AwardInput struct {
	OwnerID       int64
	CustomerID    int64
	Points        int
	Description   string
	ReferenceID   *int64
	ReferenceType string
}

type RedeemInput struct {
	OwnerID       int64
	CustomerID    int64
	Points        int
	Description   string
	ReferenceID   *int64
	ReferenceType string
}

type LoyaltyConfigInput struct {
	OwnerID             int64
	PointsPerDollar     float64
	MinRedemptionPoints int
	PointsValue         float64
	IsActive            bool
}
