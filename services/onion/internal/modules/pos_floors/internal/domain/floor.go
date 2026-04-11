package domain

import (
	"errors"
	"time"
)

var ErrFloorNotFound = errors.New("floor not found")

type Floor struct {
	ID         int64     `json:"id"`
	BusinessID int64     `json:"businessId"`
	BranchID   int64     `json:"branchId"`
	Name       string    `json:"name"`
	SortOrder  int       `json:"sortOrder"`
	IsActive   bool      `json:"isActive"`
	CreatedAt  time.Time `json:"createdAt"`
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
