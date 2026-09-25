package domain

import (
	"errors"
	"time"
)

var ErrSegmentNotFound = errors.New("segment not found")

type Segment struct {
	ID               int64      `json:"id"`
	OwnerID          int64      `json:"owner_id"`
	BusinessID       int64      `json:"business_id"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	Filters          []byte     `json:"filters"`
	CustomerCount    int        `json:"customer_count"`
	LastCalculatedAt *time.Time `json:"last_calculated_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
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

// SegmentCustomer is the customer projection returned alongside a segment -
// segments don't own customer data, this is a read-only view.
type SegmentCustomer struct {
	ID            int64
	OwnerID       int64
	FirstName     string
	LastName      string
	Email         string
	Phone         string
	DateOfBirth   *time.Time
	Gender        string
	Address       string
	City          string
	Country       string
	LoyaltyPoints int64
	TotalSpent    float64
	VisitCount    int64
	LastVisitAt   *time.Time
	LoyaltyTier   string
	Tags          []string
	Notes         string
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type SegmentDetail struct {
	Segment   Segment
	Customers []SegmentCustomer
}
