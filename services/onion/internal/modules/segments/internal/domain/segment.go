package domain

import (
	"errors"
	"time"
)

var ErrSegmentNotFound = errors.New("segment not found")

type Segment struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"ownerId"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Filters       []byte    `json:"filters"`
	CustomerCount int       `json:"customerCount"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type SegmentInput struct {
	OwnerID     int64
	Name        string
	Description string
	Filters     []byte
}

type SegmentUpdateInput struct {
	ID          int64
	OwnerID     int64
	Name        string
	Description string
	Filters     []byte
}
