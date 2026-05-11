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
	Groups    []GroupSummary
}

type InviteUserInput struct {
	Email    string
	FullName string
	Password string
	Apps     []string
	Role     Role
}

type UpdatePermissionsInput struct {
	UserID int64
	Apps   []string
	Role   Role
}

type PermissionLevel string

const (
	PermLevelViewer PermissionLevel = "viewer"
	PermLevelEditor PermissionLevel = "editor"
)

type ResourcePermission struct {
	ResourceKey     string
	PermissionLevel PermissionLevel
}

type GroupSummary struct {
	ID    int64
	Name  string
	Color string
}

type Group struct {
	ID                  int64
	OwnerID             int64
	Name                string
	Description         string
	Color               string
	ResourcePermissions []ResourcePermission
	MemberCount         int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type GroupMember struct {
	UserID   int64
	Email    string
	FullName string
	AddedAt  time.Time
}

type GroupDetail struct {
	Group
	Members []GroupMember
}

type CreateGroupInput struct {
	Name                string
	Description         string
	Color               string
	ResourcePermissions []ResourcePermission
}

type UpdateGroupInput struct {
	Name                string
	Description         string
	Color               string
	ResourcePermissions []ResourcePermission
}

type BulkInviteInput struct {
	Emails  []string
	Role    Role
	Apps    []string
	GroupID int64
}

type BulkInviteResult struct {
	Email   string
	UserID  int64
	Token   string
	Status  string
	Message string
}
