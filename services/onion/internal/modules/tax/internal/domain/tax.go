package domain

import (
	"errors"
	"time"
)

var ErrTaxConfigNotFound = errors.New("tax config not found")

type TaxConfig struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	BusinessID  int64     `json:"business_id"`
	TaxName     string    `json:"tax_name"`
	TaxRate     float64   `json:"tax_rate"`
	IsInclusive bool      `json:"is_inclusive"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TaxConfigInput struct {
	OwnerID     int64
	BusinessID  int64
	TaxName     string
	TaxRate     float64
	IsInclusive bool
	IsActive    bool
}

type PLReport struct {
	TotalRevenue  float64 `json:"total_revenue"`
	TotalExpenses float64 `json:"total_expenses"`
	GrossProfit   float64 `json:"gross_profit"`
	NetProfit     float64 `json:"net_profit"`
}
