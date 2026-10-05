package domain

import (
	"time"

	"github.com/voxire/onion/internal/platform/money"
)

type BillingSummary struct {
	Plan            string
	Status          string
	CurrentUsage    float64
	UsageLimit      float64
	NextBillingDate *time.Time
}

type Invoice struct {
	ID       int64
	Amount   money.Amount
	Currency string
	Status   string
	IssuedAt time.Time
	PaidAt   *time.Time
}
