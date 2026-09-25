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
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"owner_id"`
	BusinessID    int64     `json:"business_id"`
	ProductID     int64     `json:"product_id"`
	Name          string    `json:"name"`
	SKU           string    `json:"sku"`
	Barcode       string    `json:"barcode"`
	CostPrice     float64   `json:"cost_price"`
	SalePrice     float64   `json:"sale_price"`
	StockQuantity float64   `json:"stock_quantity"`
	Attributes    string    `json:"attributes"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
type VariantInput struct {
	ProductID     int64   `json:"product_id"`
	Name          string  `json:"name"`
	SKU           string  `json:"sku"`
	Barcode       string  `json:"barcode"`
	CostPrice     float64 `json:"cost_price"`
	SalePrice     float64 `json:"sale_price"`
	StockQuantity float64 `json:"stock_quantity"`
	Attributes    string  `json:"attributes"`
	IsActive      bool    `json:"is_active"`
}
