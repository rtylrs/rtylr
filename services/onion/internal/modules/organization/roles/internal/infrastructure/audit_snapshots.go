package infrastructure

import (
	"encoding/json"
	"time"

	"github.com/voxire/onion/internal/modules/organization/roles/internal/domain"
)

type roleAudit struct {
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

func marshalRoleAudit(value domain.Role) ([]byte, error) { return json.Marshal(roleAudit(value)) }

type createRoleInputAudit struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	AppID        string   `json:"app_id"`
	RoleKey      string   `json:"role_key"`
	Level        int32    `json:"level"`
	ParentRoleID *int64   `json:"parent_role_id,omitempty"`
	ScopeRules   []string `json:"scope_rules"`
	Permissions  []string `json:"permissions"`
}

func marshalCreateRoleInputAudit(value domain.CreateRoleInput) ([]byte, error) {
	return json.Marshal(createRoleInputAudit(value))
}

type updateRoleInputAudit struct {
	ID           int64    `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	AppID        string   `json:"app_id"`
	Level        int32    `json:"level"`
	ParentRoleID *int64   `json:"parent_role_id,omitempty"`
	ScopeRules   []string `json:"scope_rules"`
	Permissions  []string `json:"permissions"`
}

func marshalUpdateRoleInputAudit(value domain.UpdateRoleInput) ([]byte, error) {
	return json.Marshal(updateRoleInputAudit(value))
}
