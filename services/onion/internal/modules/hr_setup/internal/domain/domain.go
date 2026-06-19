package domain

import (
	"context"
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
	ErrStepNotFound   = errors.New("setup step not found")
	ErrConfigNotFound = errors.New("hr config not found")
	ErrInvalidStepKey = errors.New("invalid setup step key")
)

type Repository interface {
	GetSetupSteps(ctx context.Context, ownerID int64) ([]*SetupStep, error)
	UpsertSetupStep(ctx context.Context, ownerID int64, stepKey string, completedAt time.Time) error
	GetHRConfig(ctx context.Context, ownerID int64) (*HRConfig, error)
	UpsertHRConfig(ctx context.Context, ownerID int64, completedAt *time.Time) (*HRConfig, error)
}
