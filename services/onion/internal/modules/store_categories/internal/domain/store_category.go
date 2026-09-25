package domain

import (
	"errors"
	"time"
)

// ErrParentNotFound is returned when a category's parent_id does not reference
// a category owned by the same tenant.
var ErrParentNotFound = errors.New("parent category not found")

type StoreCategory struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	BusinessID  int64     `json:"business_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	ParentID    int64     `json:"parent_id"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
}

type StoreCategoryInput struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	ParentID    int64  `json:"parent_id"`
	SortOrder   int    `json:"sort_order"`
}
