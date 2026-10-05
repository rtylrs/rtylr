package domain

import (
	"context"
	"time"
)

type Repository interface {
	// Templates
	CreateTemplate(ctx context.Context, businessID int64, params CreateTemplateParams) (*ChecklistTemplateRecord, error)
	GetTemplate(ctx context.Context, businessID, id int64) (*ChecklistTemplateRecord, error)
	ListTemplates(ctx context.Context, businessID int64, limit, offset int32) ([]ChecklistTemplateRecord, int64, error)
	ListTemplatesFiltered(ctx context.Context, businessID int64, params ListTemplatesParams) ([]ChecklistTemplateRecord, int64, error)
	UpdateTemplate(ctx context.Context, businessID, id int64, params UpdateTemplateParams) error

	// Template Items
	CreateTemplateItem(ctx context.Context, businessID, templateID int64, params CreateTemplateItemParams) (*ChecklistItemRecord, error)
	ListTemplateItems(ctx context.Context, businessID, templateID int64) ([]ChecklistItemRecord, error)

	// Employee Checklists
	TriggerChecklist(ctx context.Context, businessID, employeeID, templateID int64, triggerDate time.Time) (*EmployeeChecklistRecord, error)
	GetEmployeeChecklist(ctx context.Context, businessID, employeeID, checklistID int64) (*EmployeeChecklistRecord, error)
	ListEmployeeChecklists(ctx context.Context, businessID, employeeID int64, limit, offset int32) ([]EmployeeChecklistRecord, int64, error)
	UpdateEmployeeChecklistStatus(ctx context.Context, businessID, checklistID int64, status string) error

	// Employee Checklist Items
	CreateEmployeeChecklistItem(ctx context.Context, businessID, employeeChecklistID, checklistItemID, assignedToEmployeeID int64, dueDate string) error
	GetChecklistItem(ctx context.Context, businessID, itemID int64) (*EmployeeChecklistItemRecord, error)
	CompleteChecklistItem(ctx context.Context, businessID, itemID, completedBy int64) error
	CompleteChecklistItemWithAttachment(ctx context.Context, businessID, itemID, completedBy int64, attachmentURL *string) error
	GetChecklistCompletionStats(ctx context.Context, businessID, checklistID int64) (total, completed int, err error)
	ListEmployeeChecklistItems(ctx context.Context, businessID, employeeChecklistID int64) ([]EmployeeChecklistItemRecord, error)

	// Employee ownership is kept in this module's repository boundary so a
	// caller-supplied employee id can never be written under a foreign tenant.
	EmployeeBelongsToBusiness(ctx context.Context, businessID, employeeID int64) (bool, error)
	GetEmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error)
}
