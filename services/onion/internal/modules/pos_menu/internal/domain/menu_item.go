package domain

import "time"

type MenuItem struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"ownerId"`
	CategoryID int64     `json:"categoryId"`
	Name       string    `json:"name"`
	SKU        string    `json:"sku"`
	SalePrice  float64   `json:"salePrice"`
	IsActive   bool      `json:"isActive"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type MenuItemInput struct {
	Name      string  `json:"name"`
	SKU       string  `json:"sku"`
	SalePrice float64 `json:"salePrice"`
}
