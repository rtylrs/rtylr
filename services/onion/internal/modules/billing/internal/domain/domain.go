package domain

import "time"

type BillingSummary struct {
	Plan            string     `json:"plan"`
	Status          string     `json:"status"`
	CurrentUsage    float64    `json:"currentUsage"`
	UsageLimit      float64    `json:"usageLimit"`
	NextBillingDate *time.Time `json:"nextBillingDate"`
}

type Invoice struct {
	ID        int64     `json:"id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	IssuedAt  time.Time `json:"issuedAt"`
	PaidAt    *time.Time `json:"paidAt"`
}
