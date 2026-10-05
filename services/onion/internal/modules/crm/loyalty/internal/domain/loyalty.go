package domain

import (
	"errors"
	"time"
)

var (
	ErrLoyaltyConfigNotFound = errors.New("loyalty config not found")
	ErrRewardNotFound        = errors.New("loyalty reward not found")
	ErrInsufficientPoints    = errors.New("insufficient loyalty points")
	ErrTransactionNotFound   = errors.New("loyalty transaction not found")
	ErrInvalidLoyaltyConfig  = errors.New("invalid loyalty config")
)

type LoyaltyTransaction struct {
	ID            int64
	OwnerID       int64
	BusinessID    int64
	CustomerID    int64
	Points        int
	Type          string
	Description   string
	ReferenceID   *int64
	ReferenceType string
	CreatedAt     time.Time
}

type LoyaltyConfig struct {
	ID                  int64
	OwnerID             int64
	BusinessID          int64
	PointsPerDollar     float64
	MinRedemptionPoints int
	PointsValue         float64
	IsActive            bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
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
	ID           int64
	OwnerID      int64
	Name         string
	Points       int
	Tier         string
	Visits       int
	JoinedAt     time.Time
	LastActiveAt time.Time
}

// LoyaltyReward is a catalog entry that members can redeem points for.
type LoyaltyReward struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	Name        string
	Description string
	Category    string
	PointsCost  int
	Stock       *int
	Icon        string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
