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

// Capabilities is the effective permission set for a POS operator. It is
// seeded from the role at PIN setup (DefaultCapabilities) and may be
// overridden per-user in the staff screen, so the stored flags — not the
// role — are the source of truth at verify time.
type Capabilities struct {
	CanVoid          bool
	CanRefund        bool
	CanOpenDrawer    bool
	CanApplyDiscount bool
	MaxDiscountPct   float64
}

// DefaultCapabilities returns the sensible starting flags for a role. Managers
// get everything; cashiers/servers get drawer + discounts; kitchen/bar get
// nothing by default. Admins editing a user can override any of these.
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

// ErrInvalidPIN is returned when a submitted PIN does not match any active
// pos_user for the given owner/display name. It is kept deliberately
// coarse-grained so that handlers cannot leak which of {display_name, pin}
// was wrong.
var ErrInvalidPIN = errors.New("invalid pin")

// POSUser is the domain representation of a pos_users row. PIN and PINHash
// are carried internally so the service layer can run the dual-path verify
// migration; they must never be serialised into an API response.
type POSUser struct {
	ID               int64
	OwnerID          int64
	UserID           int64 // team users.id this PIN belongs to
	LocationID       int64
	Name             string
	PIN              string // legacy plaintext — empty once migrated
	PINHash          string // bcrypt hash; empty for unmigrated legacy rows
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

// VerifiedPOSUser is the handler-safe projection returned after a successful
// PIN verification. It excludes PIN and PINHash but carries the effective
// capabilities so the POS client can gate manager-only actions. UserID is the
// team users.id; POSUserID is the pos_users row id.
type VerifiedPOSUser struct {
	UserID      int64
	POSUserID   int64
	LocationID  int64
	Role        POSUserRole
	DisplayName string
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

// SetPINInput is the self-service path: an operator picked from the location
// roster sets or rotates their own PIN. Role and capabilities are resolved
// server-side from the member's location access, never trusted from the client.
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
