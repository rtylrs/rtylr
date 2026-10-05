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
	ID             int64
	OwnerID        int64
	BusinessID     int64
	ProductID      int64
	Title          string
	Slug           string
	Description    string
	Images         string
	Price          float64
	CompareAtPrice float64
	IsPublished    bool
	SortOrder      int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type StoreProductInput struct {
	ProductID      int64
	Title          string
	Slug           string
	Description    string
	Images         string
	Price          float64
	CompareAtPrice float64
	IsPublished    bool
	SortOrder      int
}
