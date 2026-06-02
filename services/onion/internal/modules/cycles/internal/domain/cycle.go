package domain

import (
	"errors"
	"time"
)

var ErrCycleNotFound = errors.New("cycle not found")

// ErrWorkspaceNotFound is returned when a cycle operation references a workspace
// that does not exist or does not belong to the caller (cross-tenant guard).
var ErrWorkspaceNotFound = errors.New("workspace not found")

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
