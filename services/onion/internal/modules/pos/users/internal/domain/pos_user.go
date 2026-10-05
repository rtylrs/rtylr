package domain

import (
	"errors"
	"time"
)

type POSUserRole string

const (
	RoleManager   POSUserRole = "manager"
	RoleCashier   POSUserRole = "cashier"
	RoleServer    POSUserRole = "server"
	RoleBartender POSUserRole = "bartender"
	RoleKitchen   POSUserRole = "kitchen"
)

type Capabilities struct {
	CanVoid          bool
	CanRefund        bool
	CanOpenDrawer    bool
	CanApplyDiscount bool
	MaxDiscountPct   float64
}

func DefaultCapabilities(role POSUserRole) Capabilities {
	switch role {
	case RoleManager:
		return Capabilities{CanVoid: true, CanRefund: true, CanOpenDrawer: true, CanApplyDiscount: true, MaxDiscountPct: 100}
	case RoleCashier:
		return Capabilities{CanOpenDrawer: true, CanApplyDiscount: true, MaxDiscountPct: 15}
	case RoleServer:
		return Capabilities{CanOpenDrawer: true}
	default:
		return Capabilities{}
	}
}

var ErrInvalidPIN = errors.New("invalid pin")

var ErrUserNotFound = errors.New("pos user not found")

type LocationAccess struct {
	CanManagePOSUsers  bool
	CanManageTerminals bool
	CanViewReports     bool
}

type POSUser struct {
	ID               int64
	OwnerID          int64
	BusinessID       int64
	UserID           int64
	LocationID       int64
	Name             string
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

type VerifiedPOSUser struct {
	UserID               int64
	POSUserID            int64
	LocationID           int64
	Role                 POSUserRole
	DisplayName          string
	AuthorizationVersion int64
	Capabilities
}

type CreatePOSUserInput struct {
	UserID           int64
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

type SetPINInput struct {
	UserID     int64
	LocationID int64
	Name       string
	PIN        string
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
