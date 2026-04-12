package domain

import "time"

type BillingSummary struct {
	Plan            string     `json:"plan"`
	Status          string     `json:"status"`
	CurrentUsage    float64    `json:"current_usage"`
	UsageLimit      float64    `json:"usage_limit"`
	NextBillingDate *time.Time `json:"next_billing_date"`
}

type Invoice struct {
	ID        int64     `json:"id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	IssuedAt  time.Time `json:"issued_at"`
	PaidAt    *time.Time `json:"paid_at"`
}
