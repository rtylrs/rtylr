package domain

import "time"

type ShiftRecord struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	LocationID int64     `json:"location_id"`
	Name       string    `json:"name"`
	StartTime  string    `json:"start_time"`
	EndTime    string    `json:"end_time"`
	Color      string    `json:"color"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ShiftAssignmentRecord struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	EmployeeID int64     `json:"employee_id"`
	ShiftID    int64     `json:"shift_id"`
	Date       string    `json:"date"`
	Status     string    `json:"status"`
	Published  bool      `json:"published"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ListShiftsParams struct {
	Limit  int32
	Offset int32
}

type RosterParams struct {
	LocationID int64
	WeekStart  string
}

type ConflictInfo struct {
	EmployeeID   int64  `json:"employee_id"`
	Date         string `json:"date"`
	ConflictType string `json:"conflict_type"`
	Description  string `json:"description"`
}
