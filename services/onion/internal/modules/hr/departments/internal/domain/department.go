package domain

import (
	"errors"
	"time"
)

var ErrDepartmentNotFound = errors.New("department not found")

type Department struct {
	ID             int64
	OwnerID        int64
	BusinessID     int64
	Name           string
	Description    string
	ManagerStaffID int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
type DepartmentInput struct {
	ID             int64
	OwnerID        int64
	BusinessID     int64
	Name           string
	Description    string
	ManagerStaffID int64
}
