package domain

import "errors"

// MaxLoyaltyPoints is the largest value representable by the signed INT
// columns used for customer balances and loyalty transaction amounts.
const (
	MaxLoyaltyPoints           int64 = 1<<31 - 1
	DefaultLoyaltyTierPageSize       = 20
	MaxLoyaltyTierPageSize           = 100
)

var (
	ErrLoyaltyPointsNotPositive  = errors.New("loyalty points must be greater than zero")
	ErrLoyaltyPointsOutOfRange   = errors.New("loyalty points exceed the maximum allowed value")
	ErrLoyaltyBalanceCapacity    = errors.New("loyalty balance capacity exceeded")
	ErrLoyaltyTierPageOutOfRange = errors.New("loyalty tier page offset is too large")
)

// ValidateLoyaltyPoints applies the storage bound shared by HTTP handlers,
// generated RPCs, and application-service callers.
func ValidateLoyaltyPoints(points int64) error {
	if points <= 0 {
		return ErrLoyaltyPointsNotPositive
	}
	if points > MaxLoyaltyPoints {
		return ErrLoyaltyPointsOutOfRange
	}
	return nil
}

type LoyaltyTierPage struct {
	Tiers    []LoyaltyTier `json:"tiers"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
	HasMore  bool          `json:"has_more"`
}
