package domain

import (
	"errors"
	"time"
)

var (
	ErrTimeSlotUnavailable = errors.New("appointment time slot is unavailable")
	ErrInvalidStatus       = errors.New("invalid appointment status transition")
	ErrInvalidReference    = errors.New("appointment reference is not owned by the business")
)

type Appointment struct {
	ID            int64
	OwnerID       int64
	BusinessID    int64
	ServiceID     int64
	ResourceID    int64
	StaffID       int64
	CustomerName  string
	CustomerEmail string
	CustomerPhone string
	StartTime     time.Time
	EndTime       time.Time
	Status        string
	Notes         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type AppointmentInput struct {
	ServiceID     int64
	ResourceID    int64
	StaffID       int64
	CustomerName  string
	CustomerEmail string
	CustomerPhone string
	StartTime     string
	EndTime       string
	Status        string
	Notes         string
}
