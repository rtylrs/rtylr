package domain

import "time"

type StoreCategory struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"ownerId"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	ImageURL    string    `json:"imageUrl"`
	ParentID    int64     `json:"parentId"`
	SortOrder   int       `json:"sortOrder"`
	CreatedAt   time.Time `json:"createdAt"`
}

type StoreCategoryInput struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
	ParentID    int64  `json:"parentId"`
	SortOrder   int    `json:"sortOrder"`
}
