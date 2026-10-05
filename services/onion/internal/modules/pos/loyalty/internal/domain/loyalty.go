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
	ID              int64
	BusinessID      int64
	IsEnabled       bool
	PointsPerDollar float64
	RedemptionRate  float64
	MinRedeemPoints int32
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type UpdateLoyaltyConfigInput struct {
	BusinessID      int64
	IsEnabled       bool
	PointsPerDollar float64
	RedemptionRate  float64
	MinRedeemPoints int32
}

type LoyaltyTransaction struct {
	ID           int64
	BusinessID   int64
	CustomerID   int64
	OrderID      *int64
	Type         string
	Points       int32
	BalanceAfter int32
	Note         *string
	CreatedAt    time.Time
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
