package domain

import (
	"errors"
	"time"

	"github.com/voxire/onion/internal/platform/authorization"
)

var (
	ErrPermissionNotFound     = errors.New("user app permission not found")
	ErrUserNotFound           = errors.New("user not found")
	ErrTokenAlreadyUsed       = errors.New("invite token already used")
	ErrTokenExpired           = errors.New("invite token expired")
	ErrForbidden              = errors.New("permission denied")
	ErrSeatLimitExceeded      = errors.New("subscription seat limit exceeded")
	ErrGroupNotFound          = errors.New("group not found")
	ErrInvalidPermissionLevel = errors.New("invalid group resource permission level")
	ErrDelegationCeiling      = errors.New("authorization exceeds delegation ceiling")
)

// ValidApps is the set of recognized app identifiers.
var ValidApps = map[string]struct{}{
	"pos":      {},
	"erp":      {},
	"hr":       {},
	"crm":      {},
	"finance":  {},
	"flow":     {},
	"insights": {},
	"menu":     {},
	"recruit":  {},
}

// MyProfile is the composite /user/me payload that the team service builds
// from several sub-queries. Lives in domain so it can be cached as a single
// blob keyed by user id.
type MyProfile struct {
	ID                 int64                             `json:"id"`
	Email              string                            `json:"email"`
	FullName           string                            `json:"full_name"`
	Apps               []string                          `json:"apps"`
	Role               string                            `json:"role"`
	BusinessID         int64                             `json:"business_id"`
	BusinessName       string                            `json:"business_name"`
	IsOwner            bool                              `json:"is_owner"`
	IsConsoleAdmin     bool                              `json:"is_console_admin"`
	Onboarded          bool                              `json:"onboarded"`
	PlanSlug           string                            `json:"plan_slug"`
	SubscriptionStatus string                            `json:"subscription_status"`
	EffectiveResources map[string]string                 `json:"effective_resources"`
	EffectiveAccess    *authorization.EffectiveAccess    `json:"effective_access,omitempty"`
	ProvisioningStatus *authorization.ProvisioningStatus `json:"provisioning_status,omitempty"`
}

type Role string

const (
	RoleOwner        Role = "owner"
	RoleAdmin        Role = "admin"
	RoleOperator     Role = "operator"
	RoleBilling      Role = "billing"
	RoleMenuManager  Role = "menu_manager"
	RoleStoreManager Role = "store_manager"
	RoleStaff        Role = "staff"
	RoleViewer       Role = "viewer"
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
	Email            string
	Username         string
	FullName         string
	Password         string
	Apps             []string
	Role             Role
	ProvisioningMode string
	DashRole         string
	AppAssignments   []AppRoleAssignmentInput
	GroupIDs         []int64
}

type AppRoleAssignmentInput struct {
	AppID         string
	RoleIDs       []int64
	PrimaryRoleID *int64
	Source        string
	Status        string
	Scopes        []authorization.Scope
	GroupID       *int64
}

type AuthorizationRoleGrant struct {
	AppID       string
	Permissions []string
}

type UpdatePermissionsInput struct {
	UserID         int64
	FullName       *string
	Apps           *[]string
	Role           Role
	DashRole       string
	AppAssignments []AppRoleAssignmentInput
	GroupIDs       []int64
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
	Emails         []string
	Role           Role
	Apps           []string
	DashRole       string
	AppAssignments []AppRoleAssignmentInput
	GroupIDs       []int64
	GroupID        int64
}

type BulkInviteResult struct {
	Email   string
	UserID  int64
	Token   string
	Status  string
	Message string
}

type AccessRequest struct {
	ID             int64
	UserID         int64
	App            string
	Status         string
	RequesterName  string
	RequesterEmail string
	CreatedAt      time.Time
}
