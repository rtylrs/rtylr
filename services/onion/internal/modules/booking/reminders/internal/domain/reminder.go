package domain

import (
	"errors"
	"time"
)

// ErrAppointmentNotFound is returned when a reminder references an appointment
// that does not exist or is not owned by the requesting user.
var ErrAppointmentNotFound = errors.New("appointment not found")

type Reminder struct {
	ID            int64
	OwnerID       int64
	BusinessID    int64
	AppointmentID int64
	Type          string
	ScheduledAt   time.Time
	SentAt        time.Time
	CreatedAt     time.Time
}

type ReminderInput struct {
	AppointmentID int64
	Type          string
	ScheduledAt   string
}
