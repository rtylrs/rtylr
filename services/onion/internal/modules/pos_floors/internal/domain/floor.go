package domain

import (
	"time"
)

type Floor struct {
	ID         int64     `json:"id"`
	BusinessID int64     `json:"business_id"`
	BranchID   int64     `json:"branch_id"`
	Name       string    `json:"name"`
	SortOrder  int       `json:"sort_order"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

type FloorInput struct {
	BusinessID int64
	BranchID   int64
	Name       string
	SortOrder  int
}

type FloorUpdateInput struct {
	ID         int64
	BusinessID int64
	Name       string
	SortOrder  int
}
