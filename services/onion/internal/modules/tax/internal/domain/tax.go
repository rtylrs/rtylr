package domain

import (
	"errors"
	"time"
)

var ErrTaxConfigNotFound = errors.New("tax config not found")

type TaxConfig struct {
	ID                 int64     `json:"id"`
	OwnerID            int64     `json:"owner_id"`
	BusinessID         int64     `json:"business_id"`
	TaxName            string    `json:"tax_name"`
	TaxRateBasisPoints int64     `json:"tax_rate_basis_points"`
	IsInclusive        bool      `json:"is_inclusive"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type TaxConfigInput struct {
	OwnerID            int64
	BusinessID         int64
	TaxName            string
	TaxRateBasisPoints int64
	IsInclusive        bool
	IsActive           bool
}

type PLReport struct {
	// Monetary values are stored as integer minor units (cents for two-decimal
	// currencies) so calculations do not lose precision to floating point.
	TotalRevenueMinorUnits  int64 `json:"total_revenue_minor_units"`
	TotalExpensesMinorUnits int64 `json:"total_expenses_minor_units"`
	GrossProfitMinorUnits   int64 `json:"gross_profit_minor_units"`
	NetProfitMinorUnits     int64 `json:"net_profit_minor_units"`
}
