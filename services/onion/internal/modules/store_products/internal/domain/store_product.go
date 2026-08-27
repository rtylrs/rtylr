package domain

import (
	"errors"
	"time"
)

// ErrInvalidPrice is returned when a price or compare-at price is negative.
var ErrInvalidPrice = errors.New("invalid price: monetary values cannot be negative")

// ErrStoreProductNotFound is returned when a storefront product lookup misses.
var ErrStoreProductNotFound = errors.New("store product not found")

type StoreProduct struct {
	ID             int64     `json:"id"`
	OwnerID        int64     `json:"owner_id"`
	BusinessID     int64     `json:"business_id"`
	ProductID      int64     `json:"product_id"`
	Title          string    `json:"title"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description"`
	Images         string    `json:"images"`
	Price          float64   `json:"price"`
	CompareAtPrice float64   `json:"compare_at_price"`
	IsPublished    bool      `json:"is_published"`
	SortOrder      int       `json:"sort_order"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type StoreProductInput struct {
	ProductID      int64   `json:"product_id"`
	Title          string  `json:"title"`
	Slug           string  `json:"slug"`
	Description    string  `json:"description"`
	Images         string  `json:"images"`
	Price          float64 `json:"price"`
	CompareAtPrice float64 `json:"compare_at_price"`
	IsPublished    bool    `json:"is_published"`
	SortOrder      int     `json:"sort_order"`
}

// CentsToFloat converts integer cents to standard float amount.
func CentsToFloat(cents int64) float64 {
	return float64(cents) / 100.0
}
