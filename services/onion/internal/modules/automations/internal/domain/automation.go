package domain

import (
	"errors"
	"time"
)

var ErrAutomationNotFound = errors.New("automation not found")

type Automation struct {
	ID            int64
	OwnerID       int64
	BusinessID    int64
	Name          string
	TriggerType   string
	TriggerConfig []byte
	Actions       []byte
	IsActive      bool
	LastRunAt     *time.Time
	RunCount      int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type AutomationInput struct {
	OwnerID       int64
	BusinessID    int64
	Name          string
	TriggerType   string
	TriggerConfig []byte
	Actions       []byte
	IsActive      bool
}

type AutomationUpdateInput struct {
	ID            int64
	OwnerID       int64
	BusinessID    int64
	Name          string
	TriggerType   string
	TriggerConfig []byte
	Actions       []byte
	IsActive      bool
}
