package domain

import (
	"errors"
	"time"
)

// ErrServiceNotFound is returned when a waitlist entry references a
// booking service that does not exist or is not owned by the requester.
var ErrServiceNotFound = errors.New("booking service not found")

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
