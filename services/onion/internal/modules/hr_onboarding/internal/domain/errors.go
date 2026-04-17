package domain

import "errors"

var (
	ErrTemplateNotFound        = errors.New("checklist template not found")
	ErrChecklistNotFound       = errors.New("employee checklist not found")
	ErrChecklistItemNotFound   = errors.New("checklist item not found")
	ErrChecklistAlreadyCompleted = errors.New("checklist item is already completed")
)
