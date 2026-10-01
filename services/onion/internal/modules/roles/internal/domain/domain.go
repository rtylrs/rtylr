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
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"owner_id"`
	BusinessID   int64     `json:"business_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	AppID        string    `json:"app_id"`
	RoleKey      string    `json:"role_key"`
	Level        int32     `json:"level"`
	ParentRoleID *int64    `json:"parent_role_id,omitempty"`
	ScopeRules   []string  `json:"scope_rules"`
	Permissions  []string  `json:"permissions"`
	IsSystem     bool      `json:"is_system"`
	IsArchived   bool      `json:"is_archived"`
	Lifecycle    string    `json:"lifecycle"`
	Version      int64     `json:"version"`
	MemberCount  int32     `json:"member_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	AppID        string   `json:"app_id"`
	RoleKey      string   `json:"role_key"`
	Level        int32    `json:"level"`
	ParentRoleID *int64   `json:"parent_role_id,omitempty"`
	ScopeRules   []string `json:"scope_rules"`
	Permissions  []string `json:"permissions"`
}

type UpdateRoleInput struct {
	ID           int64    `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	AppID        string   `json:"app_id"`
	Level        int32    `json:"level"`
	ParentRoleID *int64   `json:"parent_role_id,omitempty"`
	ScopeRules   []string `json:"scope_rules"`
	Permissions  []string `json:"permissions"`
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
	Key         string `json:"key"`
	AppID       string `json:"app_id"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	IsSensitive bool   `json:"is_sensitive"`
	IsHighRisk  bool   `json:"is_high_risk"`
}
