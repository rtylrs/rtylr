package domain

import (
	"time"
)

type OneOnOne struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	BusinessID  int64     `json:"business_id"`
	ManagerID   int64     `json:"manager_id"`
	StaffID     int64     `json:"staff_id"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Notes       string    `json:"notes"`
	ActionItems string    `json:"action_items"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}
type OneOnOneInput struct {
	ManagerID   int64  `json:"manager_id"`
	StaffID     int64  `json:"staff_id"`
	ScheduledAt string `json:"scheduled_at"`
	Notes       string `json:"notes"`
	ActionItems string `json:"action_items"`
	Status      string `json:"status"`
}
