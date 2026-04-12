package domain

import "time"

type MenuItem struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	CategoryID int64     `json:"category_id"`
	Name       string    `json:"name"`
	SKU        string    `json:"sku"`
	SalePrice  float64   `json:"sale_price"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type MenuItemInput struct {
	Name      string  `json:"name"`
	SKU       string  `json:"sku"`
	SalePrice float64 `json:"sale_price"`
}
