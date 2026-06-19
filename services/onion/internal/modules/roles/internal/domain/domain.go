package domain

import "time"

type Role struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	BusinessID  int64     `json:"business_id"`
	Name        string    `json:"name"`
	Permissions []string  `json:"permissions"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateRoleInput struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type UpdateRoleInput struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

// PermissionEntry represents a single permission in the permissions matrix.
type PermissionEntry struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// AllPermissions returns the full set of available permissions.
func AllPermissions() []PermissionEntry {
	return []PermissionEntry{
		{Key: "pos.view", Label: "View POS", Description: "Access POS system", Category: "pos"},
		{Key: "pos.manage", Label: "Manage POS", Description: "Manage POS settings", Category: "pos"},
		{Key: "erp.view", Label: "View ERP", Description: "Access ERP system", Category: "erp"},
		{Key: "erp.manage", Label: "Manage ERP", Description: "Manage ERP settings", Category: "erp"},
		{Key: "hr.view", Label: "View HR", Description: "Access HR system", Category: "hr"},
		{Key: "hr.manage", Label: "Manage HR", Description: "Manage HR settings", Category: "hr"},
		{Key: "crm.view", Label: "View CRM", Description: "Access CRM system", Category: "crm"},
		{Key: "crm.manage", Label: "Manage CRM", Description: "Manage CRM settings", Category: "crm"},
		{Key: "finance.view", Label: "View Finance", Description: "Access finance system", Category: "finance"},
		{Key: "finance.manage", Label: "Manage Finance", Description: "Manage finance settings", Category: "finance"},
		{Key: "hub.users", Label: "Manage Users", Description: "Manage team users", Category: "hub"},
		{Key: "hub.settings", Label: "Manage Settings", Description: "Manage hub settings", Category: "hub"},
		{Key: "hub.billing", Label: "Manage Billing", Description: "Access billing info", Category: "hub"},
	}
}
