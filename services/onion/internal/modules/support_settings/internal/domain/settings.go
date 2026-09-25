package domain

import (
	"errors"
	"time"
)

// ErrInvalidSLAHours is returned when SLA hours are outside the
// allowed range (must be greater than 0 and at most 1000).
var ErrInvalidSLAHours = errors.New("sla hours must be between 1 and 1000")

type Settings struct {
	ID                int64     `json:"id"`
	BusinessID        int64     `json:"business_id"`
	AutoAssign        bool      `json:"auto_assign"`
	SLAHours          int       `json:"sla_hours"`
	BusinessHours     string    `json:"business_hours"`
	NotificationEmail string    `json:"notification_email"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type SettingsInput struct {
	BusinessID        int64
	AutoAssign        bool
	SLAHours          int
	BusinessHours     string
	NotificationEmail string
}
