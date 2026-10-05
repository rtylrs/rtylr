package domain

import (
	"time"
)

type Floor struct {
	ID         int64
	BusinessID int64
	BranchID   int64
	Name       string
	SortOrder  int
	IsActive   bool
	CreatedAt  time.Time
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
