package domain

import (
	"time"
)

type OneOnOne struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	ManagerID   int64
	StaffID     int64
	ScheduledAt time.Time
	Notes       string
	ActionItems string
	Status      string
	CreatedAt   time.Time
}
type OneOnOneInput struct {
	ManagerID   int64
	StaffID     int64
	ScheduledAt string
	Notes       string
	ActionItems string
	Status      string
}
