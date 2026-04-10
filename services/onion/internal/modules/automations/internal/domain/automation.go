package domain

import (
	"errors"
	"time"
)

var ErrAutomationNotFound = errors.New("automation not found")

type Automation struct {
	ID            int64      `json:"id"`
	OwnerID       int64      `json:"ownerId"`
	Name          string     `json:"name"`
	TriggerType   string     `json:"triggerType"`
	TriggerConfig []byte     `json:"triggerConfig"`
	Actions       []byte     `json:"actions"`
	IsActive      bool       `json:"isActive"`
	LastRunAt     *time.Time `json:"lastRunAt"`
	RunCount      int        `json:"runCount"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type AutomationInput struct {
	OwnerID       int64
	Name          string
	TriggerType   string
	TriggerConfig []byte
	Actions       []byte
	IsActive      bool
}

type AutomationUpdateInput struct {
	ID            int64
	OwnerID       int64
	Name          string
	TriggerType   string
	TriggerConfig []byte
	Actions       []byte
	IsActive      bool
}
