package domain

import (
	"errors"
	"time"
)

var ErrDepartmentNotFound = errors.New("department not found")

type Department struct {
	ID             int64     `json:"id"`
	OwnerID        int64     `json:"owner_id"`
	BusinessID     int64     `json:"business_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	ManagerStaffID int64     `json:"manager_staff_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
type DepartmentInput struct {
	ID             int64  `json:"id"`
	OwnerID        int64  `json:"owner_id"`
	BusinessID     int64  `json:"business_id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ManagerStaffID int64  `json:"manager_staff_id"`
}
