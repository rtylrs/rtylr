package domain

import (
	"errors"
	"time"
)

var (
	ErrConfigNotFound      = errors.New("loyalty config not found")
	ErrInsufficientPoints  = errors.New("insufficient loyalty points")
	ErrLoyaltyNotEnabled   = errors.New("loyalty program is not enabled")
	ErrMinRedeemNotReached = errors.New("minimum redeem points not reached")
	ErrOrderRequired       = errors.New("a paid order is required to earn points")
	ErrOrderNotPaid        = errors.New("order is not paid for this customer")
	ErrOrderAlreadyEarned  = errors.New("points were already earned for this order")
)

type LoyaltyConfig struct {
	ID              int64     `json:"id"`
	BusinessID      int64     `json:"business_id"`
	IsEnabled       bool      `json:"is_enabled"`
	PointsPerDollar float64   `json:"points_per_dollar"`
	RedemptionRate  float64   `json:"redemption_rate"`
	MinRedeemPoints int32     `json:"min_redeem_points"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type UpdateLoyaltyConfigInput struct {
	BusinessID      int64
	IsEnabled       bool
	PointsPerDollar float64
	RedemptionRate  float64
	MinRedeemPoints int32
}

type LoyaltyTransaction struct {
	ID           int64     `json:"id"`
	BusinessID   int64     `json:"business_id"`
	CustomerID   int64     `json:"customer_id"`
	OrderID      *int64    `json:"order_id"`
	Type         string    `json:"type"`
	Points       int32     `json:"points"`
	BalanceAfter int32     `json:"balance_after"`
	Note         *string   `json:"note"`
	CreatedAt    time.Time `json:"created_at"`
}

type EarnPointsInput struct {
	BusinessID int64
	CustomerID int64
	OrderID    *int64
	// Points is retained for wire compatibility but ignored by the service;
	// awards are calculated from the paid order total and configuration.
	Points int32
}

type RedeemPointsInput struct {
	BusinessID int64
	CustomerID int64
	Points     int32
}
