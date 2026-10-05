package domain

import "time"

type FlowWorkdayEntry struct {
	ID              int64
	WorkDayID       int64
	ProjectID       int64
	TaskID          int64
	WorkTypeID      int64
	WorkMode        string
	Description     string
	DurationMinutes int32
	IsBillable      bool
	Status          string
	ClientKey       string
	LoggedBy        int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type FlowWorkdayEntryInput struct {
	ID              int64
	OwnerID         int64
	BusinessID      int64
	EmployeeID      int64
	WorkDayID       int64
	Date            string
	ProjectID       int64
	TaskID          int64
	WorkTypeID      int64
	WorkMode        string
	Timezone        string
	Description     string
	DurationMinutes int32
	IsBillable      bool
	ClientKey       string
	LoggedBy        int64
}

type FlowWorkdayEntryOwner struct {
	WorkDayID  int64
	EmployeeID int64
	TaskID     int64
}

type FlowWorkdayMemberMinutes struct {
	EmployeeID int64
	Minutes    int64
}
