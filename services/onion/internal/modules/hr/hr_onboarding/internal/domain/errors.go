package domain

import "errors"

var (
	ErrTemplateNotFound          = errors.New("checklist template not found")
	ErrChecklistNotFound         = errors.New("employee checklist not found")
	ErrChecklistItemNotFound     = errors.New("checklist item not found")
	ErrChecklistAlreadyCompleted = errors.New("checklist item is already completed")
	ErrChecklistAccessDenied     = errors.New("checklist item access denied")
	ErrEmployeeNotFound          = errors.New("employee not found in business")
	ErrInvalidAttachmentURL      = errors.New("invalid checklist attachment URL")
)
