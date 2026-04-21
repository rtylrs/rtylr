package domain

import (
	"errors"
	"time"
)

type POSUserRole string

const (
	RoleCashier   POSUserRole = "cashier"
	RoleServer    POSUserRole = "server"
	RoleBartender POSUserRole = "bartender"
	RoleKitchen   POSUserRole = "kitchen"
)

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

// VerifiedPOSUser is the minimal, handler-safe projection returned after a
// successful PIN verification. It intentionally excludes PIN, PINHash, and
// any permission flags the caller did not ask for.
type VerifiedPOSUser struct {
	UserID      int64
	Role        POSUserRole
	DisplayName string
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
