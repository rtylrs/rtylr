package domain

import "time"

type StoreProduct struct {
	ID             int64     `json:"id"`
	OwnerID        int64     `json:"ownerId"`
	ProductID      int64     `json:"productId"`
	Title          string    `json:"title"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description"`
	Images         string    `json:"images"`
	Price          float64   `json:"price"`
	CompareAtPrice float64   `json:"compareAtPrice"`
	IsPublished    bool      `json:"isPublished"`
	SortOrder      int       `json:"sortOrder"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type StoreProductInput struct {
	ProductID      int64   `json:"productId"`
	Title          string  `json:"title"`
	Slug           string  `json:"slug"`
	Description    string  `json:"description"`
	Images         string  `json:"images"`
	Price          float64 `json:"price"`
	CompareAtPrice float64 `json:"compareAtPrice"`
	IsPublished    bool    `json:"isPublished"`
	SortOrder      int     `json:"sortOrder"`
}
