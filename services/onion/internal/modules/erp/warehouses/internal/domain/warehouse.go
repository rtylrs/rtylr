package domain

import (
	"errors"
	"time"
)

var ErrWarehouseNotFound = errors.New("warehouse not found")

type Warehouse struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	Name        string
	Address     string
	City        string
	Country     string
	ManagerName string
	Capacity    int64
	IsDefault   bool
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
type WarehouseInput struct {
	Name        string
	Address     string
	City        string
	Country     string
	ManagerName string
	Capacity    int64
	IsDefault   bool
	IsActive    bool
}
