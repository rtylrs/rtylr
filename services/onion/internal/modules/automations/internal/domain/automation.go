package domain

import (
	"errors"
	"time"
)

var ErrAutomationNotFound = errors.New("automation not found")

type Automation struct {
	ID            int64      `json:"id"`
	OwnerID       int64      `json:"owner_id"`
	BusinessID    int64      `json:"business_id"`
	Name          string     `json:"name"`
	TriggerType   string     `json:"trigger_type"`
	TriggerConfig []byte     `json:"trigger_config"`
	Actions       []byte     `json:"actions"`
	IsActive      bool       `json:"is_active"`
	LastRunAt     *time.Time `json:"last_run_at"`
	RunCount      int        `json:"run_count"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
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
