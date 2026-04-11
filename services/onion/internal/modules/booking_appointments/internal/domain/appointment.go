package domain

import "time"

type Appointment struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"ownerId"`
	ServiceID     int64     `json:"serviceId"`
	ResourceID    int64     `json:"resourceId"`
	StaffID       int64     `json:"staffId"`
	CustomerName  string    `json:"customerName"`
	CustomerEmail string    `json:"customerEmail"`
	CustomerPhone string    `json:"customerPhone"`
	StartTime     time.Time `json:"startTime"`
	EndTime       time.Time `json:"endTime"`
	Status        string    `json:"status"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type AppointmentInput struct {
	ServiceID     int64  `json:"serviceId"`
	ResourceID    int64  `json:"resourceId"`
	StaffID       int64  `json:"staffId"`
	CustomerName  string `json:"customerName"`
	CustomerEmail string `json:"customerEmail"`
	CustomerPhone string `json:"customerPhone"`
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	Status        string `json:"status"`
	Notes         string `json:"notes"`
}
