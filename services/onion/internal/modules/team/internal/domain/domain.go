package domain

import (
	"errors"
	"time"
)

var (
	ErrPermissionNotFound = errors.New("user app permission not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrTokenAlreadyUsed   = errors.New("invite token already used")
	ErrTokenExpired       = errors.New("invite token expired")
)

// ValidApps is the set of recognized app identifiers.
var ValidApps = map[string]struct{}{
	"pos":     {},
	"erp":     {},
	"hr":      {},
	"crm":     {},
	"finance": {},
	"flow":    {},
}

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleOperator Role = "operator"
	RoleStaff    Role = "staff"
	RoleViewer   Role = "viewer"
)

// UserAppPermission maps a staff user to a set of app IDs and a role for a given owner.
type UserAppPermission struct {
	ID        int64
	OwnerID   int64
	UserID    int64
	Apps      []string
	Role      Role
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// StaffUser combines basic user info with their app permissions.
type StaffUser struct {
	ID        int64
	Email     string
	FullName  string
	CreatedAt time.Time
	Apps      []string
	Role      Role
	IsActive  bool
}

type InviteUserInput struct {
	Email    string
	FullName string
	Apps     []string
	Role     Role
}

type UpdatePermissionsInput struct {
	UserID int64
	Apps   []string
	Role   Role
}
