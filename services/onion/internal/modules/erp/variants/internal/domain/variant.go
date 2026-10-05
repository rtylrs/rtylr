package domain

import (
	"errors"
	"time"
)

var (
	ErrVariantNotFound        = errors.New("variant not found")
	ErrVariantProductNotFound = errors.New("product not found in business")
)

type Variant struct {
	ID            int64
	OwnerID       int64
	BusinessID    int64
	ProductID     int64
	Name          string
	SKU           string
	Barcode       string
	CostPrice     float64
	SalePrice     float64
	StockQuantity float64
	Attributes    string
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
type VariantInput struct {
	ProductID     int64
	Name          string
	SKU           string
	Barcode       string
	CostPrice     float64
	SalePrice     float64
	StockQuantity float64
	Attributes    string
	IsActive      bool
}
