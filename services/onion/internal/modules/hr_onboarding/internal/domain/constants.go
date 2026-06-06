package domain

// Checklist trigger types.
const (
	TriggerTypeHire      = "hire"
	TriggerTypeTerminate = "terminate"
	TriggerTypeTransfer  = "transfer"
	TriggerTypePromotion = "promotion"
)

// Employee checklist statuses.
const (
	ChecklistStatusPending    = "pending"
	ChecklistStatusInProgress = "in_progress"
	ChecklistStatusCompleted  = "completed"
	ChecklistStatusCanceled   = "canceled"
)

// Checklist item statuses.
const (
	ItemStatusPending   = "pending"
	ItemStatusCompleted = "completed"
	ItemStatusSkipped   = "skipped"
)

// Assignee role values for checklist items.
const (
	AssigneeRoleHR       = "hr"
	AssigneeRoleManager  = "manager"
	AssigneeRoleEmployee = "employee"
	AssigneeRoleIT       = "it"
	AssigneeRoleFinance  = "finance"
)
