package domain

import (
	"errors"
	"time"
)

// ErrParentNotFound is returned when a category's parent_id does not reference
// a category owned by the same tenant.
var ErrParentNotFound = errors.New("parent category not found")

type StoreCategory struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	Name        string
	Slug        string
	Description string
	ImageURL    string
	ParentID    int64
	SortOrder   int
	CreatedAt   time.Time
}

type StoreCategoryInput struct {
	Name        string
	Slug        string
	Description string
	ImageURL    string
	ParentID    int64
	SortOrder   int
}
