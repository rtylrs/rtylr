package domain

import "time"

type ChecklistTemplateRecord struct {
	ID          int64                 `json:"id"`
	OwnerID     int64                 `json:"owner_id"`
	BusinessID  int64                 `json:"business_id"`
	Name        string                `json:"name"`
	TriggerType string                `json:"trigger_type"`
	RoleID      int64                 `json:"role_id"`
	Items       []ChecklistItemRecord `json:"items"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

type ChecklistItemRecord struct {
	ID                 int64     `json:"id"`
	OwnerID            int64     `json:"owner_id"`
	BusinessID         int64     `json:"business_id"`
	TemplateID         int64     `json:"template_id"`
	Title              string    `json:"title"`
	Description        string    `json:"description"`
	AssigneeRole       string    `json:"assignee_role"`
	DueOffsetDays      int32     `json:"due_offset_days"`
	RequiresAttachment bool      `json:"requires_attachment"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type EmployeeChecklistRecord struct {
	ID          int64                         `json:"id"`
	OwnerID     int64                         `json:"owner_id"`
	BusinessID  int64                         `json:"business_id"`
	EmployeeID  int64                         `json:"employee_id"`
	TemplateID  int64                         `json:"template_id"`
	Status      string                        `json:"status"`
	TriggeredAt time.Time                     `json:"triggered_at"`
	Items       []EmployeeChecklistItemRecord `json:"items"`
	CreatedAt   time.Time                     `json:"created_at"`
	UpdatedAt   time.Time                     `json:"updated_at"`
}

type EmployeeChecklistItemRecord struct {
	ID                   int64     `json:"id"`
	OwnerID              int64     `json:"owner_id"`
	BusinessID           int64     `json:"business_id"`
	EmployeeChecklistID  int64     `json:"employee_checklist_id"`
	ChecklistItemID      int64     `json:"checklist_item_id"`
	AssignedToEmployeeID int64     `json:"assigned_to_employee_id"`
	DueDate              string    `json:"due_date"`
	Status               string    `json:"status"`
	CompletedAt          time.Time `json:"completed_at"`
	CompletedBy          int64     `json:"completed_by"`
	AttachmentURL        string    `json:"attachment_url"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// EmployeeChecklistSummaryRecord is one employee checklist's progress within a
// business-wide summary. TotalItems/CompletedItems/OverdueItemsCount and the
// derived CompletionPercentage are computed server-side (the overdue count
// against the database's current time) so no caller re-derives them.
type EmployeeChecklistSummaryRecord struct {
	ChecklistID          int64     `json:"checklist_id"`
	EmployeeID           int64     `json:"employee_id"`
	TemplateID           int64     `json:"template_id"`
	Status               string    `json:"status"`
	TotalItems           int32     `json:"total_items"`
	CompletedItems       int32     `json:"completed_items"`
	CompletionPercentage int32     `json:"completion_percentage"`
	OverdueItemsCount    int32     `json:"overdue_items_count"`
	TriggeredAt          time.Time `json:"triggered_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type ListTemplatesParams struct {
	Limit  int32
	Offset int32
}

type ListEmployeeChecklistsParams struct {
	Limit  int32
	Offset int32
}

type ListBusinessEmployeeChecklistSummariesParams struct {
	Limit  int32
	Offset int32
}

type CreateTemplateParams struct {
	Name        string `json:"name"`
	TriggerType string `json:"trigger_type"`
	RoleID      int64  `json:"role_id"`
}

type UpdateTemplateParams struct {
	Name        string `json:"name"`
	TriggerType string `json:"trigger_type"`
	RoleID      int64  `json:"role_id"`
}

type CreateTemplateItemParams struct {
	Title              string `json:"title"`
	Description        string `json:"description"`
	AssigneeRole       string `json:"assignee_role"`
	DueOffsetDays      int32  `json:"due_offset_days"`
	RequiresAttachment bool   `json:"requires_attachment"`
}
