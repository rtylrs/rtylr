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
	ID            int64
	OwnerID       int64
	BusinessID    int64
	ServiceID     int64
	CustomerName  string
	CustomerEmail string
	CustomerPhone string
	PreferredDate string
	PreferredTime string
	Status        string
	CreatedAt     time.Time
}

type WaitlistInput struct {
	ServiceID     int64
	CustomerName  string
	CustomerEmail string
	CustomerPhone string
	PreferredDate string
	PreferredTime string
	Status        string
}
