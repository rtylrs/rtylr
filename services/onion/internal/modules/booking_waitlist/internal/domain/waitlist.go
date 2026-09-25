package domain

import (
	"errors"
	"time"
)

// ErrServiceNotFound is returned when a waitlist entry references a
// booking service that does not exist or is not owned by the requester.
var ErrServiceNotFound = errors.New("booking service not found")

// ErrWaitlistEntryNotFound is returned when an update matches no waitlist
// entry in the caller's business — a missing id, or another tenant's row.
var ErrWaitlistEntryNotFound = errors.New("waitlist entry not found")

// ErrPreferredDateInvalid is returned when preferred_date cannot be parsed as
// a YYYY-MM-DD date. It must not be silently stored as NULL.
var ErrPreferredDateInvalid = errors.New("preferred_date must be in YYYY-MM-DD format")

// ErrPreferredTimeInvalid is returned when preferred_time cannot be parsed as
// a 24-hour time. It must not be silently stored as NULL.
var ErrPreferredTimeInvalid = errors.New("preferred_time must be in HH:MM or HH:MM:SS format")

type WaitlistEntry struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"owner_id"`
	BusinessID    int64     `json:"business_id"`
	ServiceID     int64     `json:"service_id"`
	CustomerName  string    `json:"customer_name"`
	CustomerEmail string    `json:"customer_email"`
	CustomerPhone string    `json:"customer_phone"`
	PreferredDate string    `json:"preferred_date"`
	PreferredTime string    `json:"preferred_time"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type WaitlistInput struct {
	ServiceID     int64  `json:"service_id"`
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`
	CustomerPhone string `json:"customer_phone"`
	PreferredDate string `json:"preferred_date"`
	PreferredTime string `json:"preferred_time"`
	Status        string `json:"status"`
}
