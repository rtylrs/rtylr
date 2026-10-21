package domain

import (
	"errors"
	"time"
)

// ErrInvalidSLAHours is returned when SLA hours are outside the
// allowed range (must be greater than 0 and at most 1000).
var ErrInvalidSLAHours = errors.New("sla hours must be between 1 and 1000")

type Settings struct {
	ID                int64
	BusinessID        int64
	AutoAssign        bool
	SLAHours          int
	BusinessHours     string
	NotificationEmail string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type SettingsInput struct {
	BusinessID        int64
	AutoAssign        bool
	SLAHours          int
	BusinessHours     string
	NotificationEmail string
}
