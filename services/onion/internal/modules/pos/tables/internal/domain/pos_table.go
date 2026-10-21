package domain

import (
	"errors"
	"time"
)

var ErrTableNotFound = errors.New("table not found")

type Table struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	LocationID *int64
	Label      string
	Seats      int
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type TableInput struct {
	OwnerID    int64
	BusinessID int64
	LocationID *int64
	Label      string
	Seats      int
}

type TableUpdateInput struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	LocationID *int64
	Label      string
	Seats      int
	Status     string
}

type TableStatusInput struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	Status     string
}
