package domain

import "time"

type ChecklistTemplateRecord struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	Name        string
	TriggerType string
	RoleID      int64
	Items       []ChecklistItemRecord
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ChecklistItemRecord struct {
	ID                 int64
	OwnerID            int64
	BusinessID         int64
	TemplateID         int64
	Title              string
	Description        string
	AssigneeRole       string
	DueOffsetDays      int32
	RequiresAttachment bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type EmployeeChecklistRecord struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	EmployeeID  int64
	TemplateID  int64
	Status      string
	TriggeredAt time.Time
	Items       []EmployeeChecklistItemRecord
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type EmployeeChecklistItemRecord struct {
	ID                   int64
	OwnerID              int64
	BusinessID           int64
	EmployeeChecklistID  int64
	ChecklistItemID      int64
	AssignedToEmployeeID int64
	DueDate              string
	Status               string
	CompletedAt          time.Time
	CompletedBy          int64
	AttachmentURL        string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type ListTemplatesParams struct {
	Limit       int32
	Offset      int32
	TriggerType *string
}

type ListEmployeeChecklistsParams struct {
	Limit  int32
	Offset int32
}

type CreateTemplateParams struct {
	Name        string
	TriggerType string
	RoleID      int64
}

type UpdateTemplateParams struct {
	Name        string
	TriggerType *string
	RoleID      *int64
}

type CreateTemplateItemParams struct {
	Title              string
	Description        string
	AssigneeRole       string
	DueOffsetDays      int32
	RequiresAttachment bool
}
