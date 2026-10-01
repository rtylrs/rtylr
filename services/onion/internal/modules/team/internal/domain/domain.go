package domain

import (
	"errors"
	"time"

	"github.com/voxire/onion/internal/platform/authorization"
)

var (
	ErrPermissionNotFound = errors.New("user app permission not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrTokenAlreadyUsed   = errors.New("invite token already used")
	ErrTokenExpired       = errors.New("invite token expired")
	ErrForbidden          = errors.New("permission denied")
	ErrSeatLimitExceeded  = errors.New("subscription seat limit exceeded")
	ErrGroupNotFound      = errors.New("group not found")
	ErrDelegationCeiling  = errors.New("authorization exceeds delegation ceiling")
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
	BusinessID         int64                             `json:"business_id"`
	BusinessName       string                            `json:"business_name"`
	IsConsoleAdmin     bool                              `json:"is_console_admin"`
	Onboarded          bool                              `json:"onboarded"`
	PlanSlug           string                            `json:"plan_slug"`
	SubscriptionStatus string                            `json:"subscription_status"`
	EffectiveAccess    *authorization.EffectiveAccess    `json:"effective_access,omitempty"`
	ProvisioningStatus *authorization.ProvisioningStatus `json:"provisioning_status,omitempty"`
}

// StaffUser combines basic user info with their app permissions.
type StaffUser struct {
	ID                           int64
	MemberType                   string
	Email                        string
	FullName                     string
	CreatedAt                    time.Time
	LicensedAppIDs               []string
	PermissionProfileAssignments []authorization.RoleAssignment
	EffectiveAccess              *authorization.EffectiveAccess
	IsActive                     bool
	Groups                       []GroupSummary
}

// InviteToken is the authoritative binding carried by an outstanding
// invitation: the member, the owner who issued it, and the business it granted
// access in. Accepting must resolve the tenant from this row, never from the
// invitee's own default business.
type InviteToken struct {
	UserID     int64
	OwnerID    int64
	BusinessID int64
	ExpiresAt  time.Time
}

type InviteUserInput struct {
	Email                        string
	MemberType                   string
	LicensedAppIDs               []string
	PermissionProfileAssignments []AppRoleAssignmentInput
	Username                     string
	FullName                     string
	Password                     string
	ProvisioningMode             string
	GroupIDs                     []int64
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
	UserID                       int64
	MemberType                   string
	LicensedAppIDs               *[]string
	PermissionProfileAssignments *[]AppRoleAssignmentInput
	FullName                     *string
	GroupIDs                     *[]int64
}

type GroupSummary struct {
	ID    int64
	Name  string
	Color string
}

type Group struct {
	ID                           int64
	OwnerID                      int64
	BusinessID                   int64
	Name                         string
	Description                  string
	Color                        string
	LicensedAppIDs               []string
	PermissionProfileAssignments []authorization.RoleAssignment
	MemberCount                  int
	CreatedAt                    time.Time
	UpdatedAt                    time.Time
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
	BusinessID                   int64
	Name                         string
	Description                  string
	Color                        string
	LicensedAppIDs               []string
	PermissionProfileAssignments []AppRoleAssignmentInput
}

type UpdateGroupInput struct {
	BusinessID                   int64
	Name                         string
	Description                  string
	Color                        string
	LicensedAppIDs               []string
	PermissionProfileAssignments []AppRoleAssignmentInput
}

type BulkInviteInput struct {
	Emails                       []string
	MemberType                   string
	LicensedAppIDs               []string
	PermissionProfileAssignments []AppRoleAssignmentInput
	GroupIDs                     []int64
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
