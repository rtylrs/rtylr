package domain

import (
	"errors"
	"time"

	"github.com/voxire/onion/internal/platform/money"
)

var (
	ErrAddonNotFound    = errors.New("addon not found")
	ErrAddonNotEntitled = errors.New("addon entitlement required")
)

type Addon struct {
	ID               string
	Name             string
	Description      string
	Icon             string
	Price            money.Amount
	IsFree           bool
	IsDefaultEnabled bool
	SortOrder        int
}

type BusinessAddon struct {
	AddonID     string
	Name        string
	Description string
	Icon        string
	Price       money.Amount
	IsFree      bool
	IsEnabled   bool
	EnabledAt   *time.Time
	DisabledAt  *time.Time
}
