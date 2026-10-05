package domain

import "time"

type MenuItem struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	CategoryID  int64
	Category    string
	Name        string
	SKU         string
	Barcode     string
	SalePrice   float64
	CostPrice   float64
	TaxRate     float64
	Description string
	ImageURL    string
	Stock       float64
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type MenuItemInput struct {
	Name      string
	SKU       string
	SalePrice float64
}

type MenuItemUpdateInput struct {
	Name      string
	SalePrice float64
	Stock     float64
	IsActive  bool
}
