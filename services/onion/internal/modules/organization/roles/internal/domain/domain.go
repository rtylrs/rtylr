package domain

import (
	"errors"
	"time"
)

var (
	ErrRoleNotFound                   = errors.New("role not found")
	ErrSystemRoleMutation             = errors.New("system roles cannot be mutated")
	ErrRoleArchived                   = errors.New("role is archived")
	ErrInvalidPermission              = errors.New("invalid permission")
	ErrInvalidAssignmentTarget        = errors.New("exactly one assignment target is required")
	ErrInvalidScope                   = errors.New("scope is not owned by the business")
	ErrDelegationCeiling              = errors.New("authorization exceeds delegation ceiling")
	ErrAuthorizationSchemaUnavailable = errors.New("normalized authorization schema unavailable")
)

type Role struct {
	ID           int64
	OwnerID      int64
	BusinessID   int64
	Name         string
	Description  string
	AppID        string
	RoleKey      string
	Level        int32
	ParentRoleID *int64
	ScopeRules   []string
	Permissions  []string
	IsSystem     bool
	IsArchived   bool
	Lifecycle    string
	Version      int64
	MemberCount  int32
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RoleFilter struct {
	AppID    string
	Page     int64
	PageSize int64
}

type RolePage struct {
	Roles    []Role
	Page     int64
	PageSize int64
	Total    int64
}

type CreateRoleInput struct {
	Name         string
	Description  string
	AppID        string
	RoleKey      string
	Level        int32
	ParentRoleID *int64
	ScopeRules   []string
	Permissions  []string
}

type UpdateRoleInput struct {
	ID           int64
	Name         string
	Description  string
	AppID        string
	Level        int32
	ParentRoleID *int64
	ScopeRules   []string
	Permissions  []string
}

type RoleAssignment struct {
	ID        int64
	RoleID    int64
	AppID     string
	UserID    *int64
	GroupID   *int64
	Status    string
	Source    string
	IsPrimary bool
	Scopes    []string
	GrantedAt time.Time
}

type AssignmentFilter struct {
	RoleID   *int64
	AppID    string
	Page     int64
	PageSize int64
}

type RoleAssignmentPage struct {
	Assignments []RoleAssignment
	Page        int64
	PageSize    int64
	Total       int64
}

type AssignRoleInput struct {
	RoleID    int64
	AppID     string
	UserID    *int64
	GroupID   *int64
	IsPrimary bool
	Scopes    []string
}

type AuditEvent struct {
	ID            int64
	Action        string
	AppID         string
	RoleID        *int64
	ActorUserID   *int64
	TargetUserID  *int64
	TargetGroupID *int64
	BeforeJSON    string
	AfterJSON     string
	Reason        string
	CreatedAt     time.Time
}

type AuditFilter struct {
	RoleID   *int64
	Page     int64
	PageSize int64
}

type AuditPage struct {
	Events   []AuditEvent
	Page     int64
	PageSize int64
	Total    int64
}

type PermissionImpact struct {
	AddedPermissions    []string
	RemovedPermissions  []string
	AffectedMemberCount int32
}

// PermissionEntry represents a single permission in the permissions matrix.
type PermissionEntry struct {
	Key         string
	AppID       string
	Resource    string
	Action      string
	IsSensitive bool
	IsHighRisk  bool
}
