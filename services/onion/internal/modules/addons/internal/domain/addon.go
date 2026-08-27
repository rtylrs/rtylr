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
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	Description      string       `json:"description"`
	Icon             string       `json:"icon"`
	Price            money.Amount `json:"price"`
	IsFree           bool         `json:"is_free"`
	IsDefaultEnabled bool         `json:"is_default_enabled"`
	SortOrder        int          `json:"sort_order"`
}

type BusinessAddon struct {
	AddonID     string       `json:"addon_id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Icon        string       `json:"icon"`
	Price       money.Amount `json:"price"`
	IsFree      bool         `json:"is_free"`
	IsEnabled   bool         `json:"is_enabled"`
	EnabledAt   *time.Time   `json:"enabled_at,omitempty"`
	DisabledAt  *time.Time   `json:"disabled_at,omitempty"`
}
