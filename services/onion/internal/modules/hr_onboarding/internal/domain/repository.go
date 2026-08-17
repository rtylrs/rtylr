package domain

import (
	"context"
	"time"
)

type Repository interface {
	// Templates
	CreateTemplate(ctx context.Context, ownerID int64, params CreateTemplateParams) (*ChecklistTemplateRecord, error)
	GetTemplate(ctx context.Context, ownerID, id int64) (*ChecklistTemplateRecord, error)
	ListTemplates(ctx context.Context, ownerID int64, limit, offset int32) ([]ChecklistTemplateRecord, int64, error)
	UpdateTemplate(ctx context.Context, ownerID, id int64, params UpdateTemplateParams) error

	// Template Items
	CreateTemplateItem(ctx context.Context, ownerID, templateID int64, params CreateTemplateItemParams) (*ChecklistItemRecord, error)
	ListTemplateItems(ctx context.Context, ownerID, templateID int64) ([]ChecklistItemRecord, error)

	// Employee Checklists
	TriggerChecklist(ctx context.Context, ownerID, employeeID, templateID int64, triggerDate time.Time) (*EmployeeChecklistRecord, error)
	GetEmployeeChecklist(ctx context.Context, ownerID, employeeID, checklistID int64) (*EmployeeChecklistRecord, error)
	ListEmployeeChecklists(ctx context.Context, ownerID, employeeID int64, limit, offset int32) ([]EmployeeChecklistRecord, int64, error)
	ListBusinessEmployeeChecklistSummaries(ctx context.Context, ownerID int64, limit, offset int32) ([]EmployeeChecklistSummaryRecord, int64, error)
	UpdateEmployeeChecklistStatus(ctx context.Context, ownerID, checklistID int64, status string) error

	// Employee Checklist Items
	CreateEmployeeChecklistItem(ctx context.Context, ownerID, employeeChecklistID, checklistItemID, assignedToEmployeeID int64, dueDate string) error
	GetChecklistItem(ctx context.Context, ownerID, itemID int64) (*EmployeeChecklistItemRecord, error)
	CompleteChecklistItem(ctx context.Context, ownerID, itemID, completedBy int64) error
	GetChecklistCompletionStats(ctx context.Context, ownerID, checklistID int64) (total, completed int, err error)
	ListEmployeeChecklistItems(ctx context.Context, ownerID, employeeChecklistID int64) ([]EmployeeChecklistItemRecord, error)
}
