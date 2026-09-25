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
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"owner_id"`
	BusinessID    int64     `json:"business_id"`
	ServiceID     int64     `json:"service_id"`
	ResourceID    int64     `json:"resource_id"`
	StaffID       int64     `json:"staff_id"`
	CustomerName  string    `json:"customer_name"`
	CustomerEmail string    `json:"customer_email"`
	CustomerPhone string    `json:"customer_phone"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Status        string    `json:"status"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AppointmentInput struct {
	ServiceID     int64  `json:"service_id"`
	ResourceID    int64  `json:"resource_id"`
	StaffID       int64  `json:"staff_id"`
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`
	CustomerPhone string `json:"customer_phone"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Status        string `json:"status"`
	Notes         string `json:"notes"`
}
