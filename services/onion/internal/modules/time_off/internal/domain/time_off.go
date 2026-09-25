package domain

import (
	"time"
)

type TimeOffRequest struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	StaffID    int64     `json:"staff_id"`
	Type       string    `json:"type"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	DaysCount  float64   `json:"days_count"`
	Status     string    `json:"status"`
	Reason     string    `json:"reason"`
	ApprovedBy int64     `json:"approved_by"`
	ApprovedAt time.Time `json:"approved_at"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
type TimeOffInput struct {
	OwnerID    int64   `json:"owner_id"`
	BusinessID int64   `json:"business_id"`
	StaffID    int64   `json:"staff_id"`
	Type       string  `json:"type"`
	StartDate  string  `json:"start_date"`
	EndDate    string  `json:"end_date"`
	DaysCount  float64 `json:"days_count"`
	Reason     string  `json:"reason"`
	Notes      string  `json:"notes"`
}

type TimeOffApprovalInput struct {
	ID         int64  `json:"id"`
	Status     string `json:"status"`
	ApprovedBy int64  `json:"approved_by"`
	Notes      string `json:"notes"`
}
