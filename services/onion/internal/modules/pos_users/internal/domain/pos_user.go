package domain

import "time"

type POSUserRole string

const (
	RoleCashier   POSUserRole = "cashier"
	RoleServer    POSUserRole = "server"
	RoleBartender POSUserRole = "bartender"
	RoleKitchen   POSUserRole = "kitchen"
)

type POSUser struct {
	ID               int64
	BusinessID       int64
	LocationID       int64
	Name             string
	PIN              string
	Role             POSUserRole
	IsActive         bool
	HourlyRate       float64
	MaxDiscountPct   float64
	CanVoid          bool
	CanRefund        bool
	CanOpenDrawer    bool
	CanApplyDiscount bool
	CreatedBy        int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type CreatePOSUserInput struct {
	LocationID       int64
	Name             string
	PIN              string
	Role             POSUserRole
	HourlyRate       float64
	MaxDiscountPct   float64
	CanVoid          bool
	CanRefund        bool
	CanOpenDrawer    bool
	CanApplyDiscount bool
}

type UpdatePOSUserInput struct {
	Name             string
	PIN              string
	Role             POSUserRole
	IsActive         bool
	HourlyRate       float64
	MaxDiscountPct   float64
	CanVoid          bool
	CanRefund        bool
	CanOpenDrawer    bool
	CanApplyDiscount bool
}
