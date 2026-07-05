package domain

import (
	"errors"
	"time"
)

var ErrSegmentNotFound = errors.New("segment not found")

type Segment struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"owner_id"`
	BusinessID    int64     `json:"business_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Filters       []byte    `json:"filters"`
	CustomerCount int       `json:"customer_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SegmentInput struct {
	OwnerID     int64
	BusinessID  int64
	Name        string
	Description string
	Filters     []byte
}

type SegmentUpdateInput struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	Name        string
	Description string
	Filters     []byte
}

// SegmentCustomer is the minimal customer projection returned alongside a
// segment - segments don't own customer data, this is a read-only view.
type SegmentCustomer struct {
	ID        int64
	FirstName string
	LastName  string
	Email     string
	Phone     string
}

type SegmentDetail struct {
	Segment   Segment
	Customers []SegmentCustomer
}
