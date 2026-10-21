package domain

import (
	"errors"
	"time"
)

type SetupStep struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	StepKey     string
	CompletedAt *time.Time
}

type HRConfig struct {
	ID               int64
	OwnerID          int64
	BusinessID       int64
	SetupCompletedAt *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

var (
	ErrConfigNotFound = errors.New("hr config not found")
	ErrInvalidStepKey = errors.New("invalid setup step key")
)
