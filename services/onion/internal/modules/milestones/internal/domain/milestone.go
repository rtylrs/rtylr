package domain

import (
	"errors"
	"time"
)

var (
	ErrMilestoneNotFound = errors.New("milestone not found")
)

type Milestone struct {
	ID          int64
	OwnerID     int64
	ProjectID   int64
	Name        string
	Description string
	TargetDate  *time.Time
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type MilestoneInput struct {
	OwnerID     int64
	ProjectID   int64
	Name        string
	Description string
	TargetDate  *time.Time
}

type MilestoneUpdateInput struct {
	ID          int64
	OwnerID     int64
	Name        string
	Description string
	TargetDate  *time.Time
	Status      string
}
