package hubapp

import (
	"context"
	"errors"
	"time"
)

var (
	ErrAddonNotFound    = errors.New("addon not found")
	ErrAddonNotEntitled = errors.New("addon entitlement required")
)

type BusinessAddon struct {
	AddonID     string
	Name        string
	Description string
	Icon        string
	Price       float64
	IsFree      bool
	IsEnabled   bool
	EnabledAt   *time.Time
	DisabledAt  *time.Time
}

type AddonProvider interface {
	ListForBusiness(ctx context.Context, businessID int64) ([]BusinessAddon, error)
	Enable(ctx context.Context, businessID int64, addonID string) error
	Disable(ctx context.Context, businessID int64, addonID string) error
	IsEnabled(ctx context.Context, businessID int64, addonID string) bool
}
