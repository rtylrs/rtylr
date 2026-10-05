package domain

import (
	"errors"
	"net/url"
	"strings"
)

var (
	ErrEmptyTemplateName      = errors.New("checklist template name is required")
	ErrInvalidTriggerType     = errors.New("invalid trigger type")
	ErrEmptyItemTitle         = errors.New("checklist item title is required")
	ErrInvalidChecklistStatus = errors.New("invalid checklist status")
	ErrInvalidItemStatus      = errors.New("invalid checklist item status")
	ErrInvalidAssigneeRole    = errors.New("invalid assignee role")
)

// IsValidTriggerType returns true when the trigger type is recognised.
func IsValidTriggerType(t string) bool {
	switch t {
	case TriggerTypeHire, TriggerTypeTerminate, TriggerTypeTermination, TriggerTypeTransfer,
		TriggerTypePromotion, TriggerTypeOnboarding, TriggerTypeOffboarding:
		return true
	}
	return false
}

// ValidateAttachmentURL keeps completion payloads to ordinary web URLs. The
// value is persisted as an attachment reference; accepting javascript or other
// schemes would make the returned payload unsafe for clients to render.
func ValidateAttachmentURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 2048 {
		return ErrInvalidAttachmentURL
	}

	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return ErrInvalidAttachmentURL
	}
	return nil
}

// IsValidChecklistStatus returns true when the status is recognised.
func IsValidChecklistStatus(s string) bool {
	switch s {
	case ChecklistStatusPending, ChecklistStatusInProgress, ChecklistStatusCompleted, ChecklistStatusCanceled:
		return true
	}
	return false
}

// IsValidItemStatus returns true when the item status is recognised.
func IsValidItemStatus(s string) bool {
	switch s {
	case ItemStatusPending, ItemStatusCompleted, ItemStatusCompletedLegacy, ItemStatusSkipped:
		return true
	}
	return false
}

// IsValidAssigneeRole returns true when the assignee role is recognised.
func IsValidAssigneeRole(r string) bool {
	switch r {
	case AssigneeRoleHR, AssigneeRoleManager, AssigneeRoleEmployee, AssigneeRoleIT, AssigneeRoleFinance:
		return true
	}
	return false
}

// ValidateCreateTemplateParams checks required fields for creating a checklist template.
func ValidateCreateTemplateParams(p CreateTemplateParams) error {
	if strings.TrimSpace(p.Name) == "" {
		return ErrEmptyTemplateName
	}
	if p.TriggerType != "" && !IsValidTriggerType(p.TriggerType) {
		return ErrInvalidTriggerType
	}
	return nil
}

// ValidateCreateTemplateItemParams checks required fields for a checklist item.
func ValidateCreateTemplateItemParams(p CreateTemplateItemParams) error {
	if strings.TrimSpace(p.Title) == "" {
		return ErrEmptyItemTitle
	}
	if p.AssigneeRole != "" && !IsValidAssigneeRole(p.AssigneeRole) {
		return ErrInvalidAssigneeRole
	}
	return nil
}
