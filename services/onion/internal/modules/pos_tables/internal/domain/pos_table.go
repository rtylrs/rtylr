package domain

import (
	"errors"
	"time"
)

var ErrTableNotFound = errors.New("table not found")

type Table struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	LocationID *int64    `json:"location_id"`
	Label      string    `json:"label"`
	Seats      int       `json:"seats"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
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
