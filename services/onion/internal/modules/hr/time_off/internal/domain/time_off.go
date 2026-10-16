package domain

import (
	"errors"
	"time"
)

var ErrTimeOffNotPending = errors.New("time off request not found or already decided")

var ErrStaffNotFound = errors.New("staff member not found in this business")

type TimeOffRequest struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	StaffID    int64
	StaffName  string
	Type       string
	StartDate  time.Time
	EndDate    time.Time
	DaysCount  float64
	Status     string
	Reason     string
	ApprovedBy int64
	ApprovedAt time.Time
	Notes      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type TimeOffInput struct {
	OwnerID    int64
	BusinessID int64
	StaffID    int64
	Type       string
	StartDate  string
	EndDate    string
	DaysCount  float64
	Reason     string
	Notes      string
}

type TimeOffApprovalInput struct {
	ID         int64
	Status     string
	ApprovedBy int64
	Notes      string
}
