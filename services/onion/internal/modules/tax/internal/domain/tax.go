package domain

import (
	"errors"
	"time"
)

var ErrTaxConfigNotFound = errors.New("tax config not found")

type TaxConfig struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"ownerId"`
	TaxName     string    `json:"taxName"`
	TaxRate     float64   `json:"taxRate"`
	IsInclusive bool      `json:"isInclusive"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type TaxConfigInput struct {
	OwnerID     int64
	TaxName     string
	TaxRate     float64
	IsInclusive bool
	IsActive    bool
}

type PLReport struct {
	TotalRevenue  float64 `json:"totalRevenue"`
	TotalExpenses float64 `json:"totalExpenses"`
	GrossProfit   float64 `json:"grossProfit"`
	NetProfit     float64 `json:"netProfit"`
}
