package domain

import (
	"errors"
	"time"
)

var (
	ErrConfigNotFound       = errors.New("loyalty config not found")
	ErrInsufficientPoints   = errors.New("insufficient loyalty points")
	ErrLoyaltyNotEnabled    = errors.New("loyalty program is not enabled")
	ErrMinRedeemNotReached  = errors.New("minimum redeem points not reached")
)

type LoyaltyConfig struct {
	ID              int64     `json:"id"`
	BusinessID      int64     `json:"businessId"`
	IsEnabled       bool      `json:"isEnabled"`
	PointsPerDollar float64   `json:"pointsPerDollar"`
	RedemptionRate  float64   `json:"redemptionRate"`
	MinRedeemPoints int32     `json:"minRedeemPoints"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
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
	BusinessID   int64     `json:"businessId"`
	CustomerID   int64     `json:"customerId"`
	OrderID      *int64    `json:"orderId"`
	Type         string    `json:"type"`
	Points       int32     `json:"points"`
	BalanceAfter int32     `json:"balanceAfter"`
	Note         *string   `json:"note"`
	CreatedAt    time.Time `json:"createdAt"`
}

type EarnPointsInput struct {
	BusinessID int64
	CustomerID int64
	OrderID    *int64
	Points     int32
}

type RedeemPointsInput struct {
	BusinessID int64
	CustomerID int64
	Points     int32
}
