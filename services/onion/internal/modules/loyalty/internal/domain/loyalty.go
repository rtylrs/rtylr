package domain

import (
	"errors"
	"time"
)

var (
	ErrLoyaltyConfigNotFound = errors.New("loyalty config not found")
	ErrInsufficientPoints    = errors.New("insufficient loyalty points")
	ErrTransactionNotFound   = errors.New("loyalty transaction not found")
	ErrInvalidLoyaltyConfig  = errors.New("invalid loyalty config")
)

type LoyaltyTransaction struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"owner_id"`
	BusinessID    int64     `json:"business_id"`
	CustomerID    int64     `json:"customer_id"`
	Points        int       `json:"points"`
	Type          string    `json:"type"`
	Description   string    `json:"description"`
	ReferenceID   *int64    `json:"reference_id"`
	ReferenceType string    `json:"reference_type"`
	CreatedAt     time.Time `json:"created_at"`
}

type LoyaltyConfig struct {
	ID                  int64     `json:"id"`
	OwnerID             int64     `json:"owner_id"`
	BusinessID          int64     `json:"business_id"`
	PointsPerDollar     float64   `json:"points_per_dollar"`
	MinRedemptionPoints int       `json:"min_redemption_points"`
	PointsValue         float64   `json:"points_value"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type AwardInput struct {
	OwnerID       int64
	BusinessID    int64
	CustomerID    int64
	Points        int
	Description   string
	ReferenceID   *int64
	ReferenceType string
}

type RedeemInput struct {
	OwnerID       int64
	BusinessID    int64
	CustomerID    int64
	Points        int
	Description   string
	ReferenceID   *int64
	ReferenceType string
}

type LoyaltyConfigInput struct {
	OwnerID             int64
	BusinessID          int64
	PointsPerDollar     float64
	MinRedemptionPoints int
	PointsValue         float64
	IsActive            bool
}

// LoyaltyMember is a customer enrolled in the program — a projection
// over crm_customers filtered to those with loyalty engagement.
type LoyaltyMember struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"owner_id"`
	Name         string    `json:"name"`
	Points       int       `json:"points"`
	Tier         string    `json:"tier"`
	Visits       int       `json:"visits"`
	JoinedAt     time.Time `json:"joined_at"`
	LastActiveAt time.Time `json:"last_active_at"`
}

// LoyaltyReward is a catalog entry that members can redeem points for.
type LoyaltyReward struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	BusinessID  int64     `json:"business_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	PointsCost  int       `json:"points_cost"`
	Stock       *int      `json:"stock"`
	Icon        string    `json:"icon"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
