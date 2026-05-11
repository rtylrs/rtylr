package domain

import (
	"errors"
	"time"
)

var ErrCycleNotFound = errors.New("cycle not found")

type Cycle struct {
	ID          int64
	OwnerID     int64
	WorkspaceID int64
	Name        string
	StartDate   *time.Time
	EndDate     *time.Time
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CycleInput struct {
	OwnerID     int64
	WorkspaceID int64
	Name        string
	StartDate   *time.Time
	EndDate     *time.Time
}

type CycleUpdateInput struct {
	ID        int64
	OwnerID   int64
	Name      string
	StartDate *time.Time
	EndDate   *time.Time
	Status    string
}
