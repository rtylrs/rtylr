package domain

import "time"

type MenuItem struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	BusinessID  int64     `json:"business_id"`
	CategoryID  int64     `json:"category_id"`
	Category    string    `json:"category"`
	Name        string    `json:"name"`
	SKU         string    `json:"sku"`
	Barcode     string    `json:"barcode"`
	SalePrice   float64   `json:"sale_price"`
	CostPrice   float64   `json:"cost_price"`
	TaxRate     float64   `json:"tax_rate"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	Stock       float64   `json:"stock"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type MenuItemInput struct {
	Name      string  `json:"name"`
	SKU       string  `json:"sku"`
	SalePrice float64 `json:"sale_price"`
}

type MenuItemUpdateInput struct {
	Name      string  `json:"name"`
	SalePrice float64 `json:"sale_price"`
	Stock     float64 `json:"stock"`
	IsActive  bool    `json:"is_active"`
}
