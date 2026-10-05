package infrastructure

import (
	"time"
)

type warehouseInputSnapshot struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	City        string `json:"city"`
	Country     string `json:"country"`
	ManagerName string `json:"manager_name"`
	Capacity    int64  `json:"capacity"`
	IsDefault   bool   `json:"is_default"`
	IsActive    bool   `json:"is_active"`
}

type warehouseSnapshot struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	BusinessID  int64     `json:"business_id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	City        string    `json:"city"`
	Country     string    `json:"country"`
	ManagerName string    `json:"manager_name"`
	Capacity    int64     `json:"capacity"`
	IsDefault   bool      `json:"is_default"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
