package domain

import (
	"errors"
	"time"
)

var ErrTaxConfigNotFound = errors.New("tax config not found")

type TaxConfig struct {
	ID                 int64
	OwnerID            int64
	BusinessID         int64
	TaxName            string
	TaxRateBasisPoints int64
	IsInclusive        bool
	IsActive           bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
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
	TotalRevenueMinorUnits  int64
	TotalExpensesMinorUnits int64
	GrossProfitMinorUnits   int64
	NetProfitMinorUnits     int64
}
